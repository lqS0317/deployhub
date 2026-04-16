package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/service/cluster"
	"deployhub/internal/service/notification"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ProgressBroadcaster WS Hub 接口（由 ws 包提供）
type ProgressBroadcaster interface {
	Broadcast(room string, payload []byte)
}

// RolloutProgress 滚动更新进度消息
type RolloutProgress struct {
	DeploymentID  uint   `json:"deployment_id"`
	Status        string `json:"status"`
	Replicas      int32  `json:"replicas"`
	ReadyReplicas int32  `json:"ready_replicas"`
	Message       string `json:"message"`
}

// RolloutWatcher 监听 K8s Deployment 的滚动更新进度
type RolloutWatcher struct {
	clientPool      *cluster.ClientsetPool
	deploySvc       *DeployService
	broadcaster     ProgressBroadcaster
	notifDispatcher *notification.Dispatcher
}

// NewRolloutWatcher 创建滚动更新监听器
func NewRolloutWatcher(clientPool *cluster.ClientsetPool, deploySvc *DeployService, broadcaster ProgressBroadcaster, notifDispatcher *notification.Dispatcher) *RolloutWatcher {
	return &RolloutWatcher{
		clientPool:      clientPool,
		deploySvc:       deploySvc,
		broadcaster:     broadcaster,
		notifDispatcher: notifDispatcher,
	}
}

// Watch 启动异步 goroutine 监听指定部署的滚动更新进度
func (w *RolloutWatcher) Watch(deployment *model.Deployment, service *model.Service) {
	go w.watchRollout(deployment, service)
}

func (w *RolloutWatcher) watchRollout(deployment *model.Deployment, service *model.Service) {
	cs, err := w.clientPool.GetClientset(deployment.ClusterID)
	if err != nil {
		log.Printf("[RolloutWatcher] 获取集群客户端失败, deploymentID=%d: %v", deployment.ID, err)
		w.markFailed(deployment.ID, fmt.Sprintf("获取集群客户端失败: %v", err))
		return
	}

	room := fmt.Sprintf("deployment:%d", deployment.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.markFailed(deployment.ID, "部署超时（10分钟）")
			w.broadcastProgress(room, deployment.ID, model.DeployStatusFailed, 0, 0, "部署超时")
			return

		case <-ticker.C:
			var desired, ready, updated int32

			wt := resolveWorkloadType(deployment, service)
			if wt == "statefulset" {
				sts, err := cs.AppsV1().StatefulSets(deployment.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
				if err != nil {
					log.Printf("[RolloutWatcher] 查询 K8s StatefulSet 失败, deploymentID=%d: %v", deployment.ID, err)
					continue
				}
				desired = sts.Status.Replicas
				ready = sts.Status.ReadyReplicas
				updated = sts.Status.UpdatedReplicas
			} else {
				k8sDep, err := cs.AppsV1().Deployments(deployment.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
				if err != nil {
					log.Printf("[RolloutWatcher] 查询 K8s Deployment 失败, deploymentID=%d: %v", deployment.ID, err)
					continue
				}
				desired = k8sDep.Status.Replicas
				ready = k8sDep.Status.ReadyReplicas
				updated = k8sDep.Status.UpdatedReplicas
			}

			workloadLabel := "Deployment"
			if wt == "statefulset" {
				workloadLabel = "StatefulSet"
			}

			w.broadcastProgress(room, deployment.ID, model.DeployStatusDeploying, desired, ready,
				fmt.Sprintf("%s 更新中: %d/%d 就绪, %d 已更新", workloadLabel, ready, desired, updated))

			if desired > 0 && ready >= desired && updated >= desired {
				// 滚动更新完成，进入 Pod 健康检查阶段
				w.markPodChecking(deployment.ID)
				w.broadcastProgress(room, deployment.ID, model.DeployStatusPodChecking, desired, ready,
					"滚动更新完成，开始 Pod 健康检查（60秒观察期）")
				w.podHealthCheck(ctx, cs, deployment, service, room)
				return
			}
		}
	}
}

const (
	podCheckInterval = 5 * time.Second
	podCheckDuration = 60 * time.Second
)

// podLabelSelector 根据部署类型选择正确的 Pod 标签选择器
func podLabelSelector(deployment *model.Deployment, service *model.Service) string {
	if deployment.DeployType == "helm" {
		releaseName := deployment.HelmReleaseName
		if releaseName == "" && service != nil {
			releaseName = service.HelmReleaseName
		}
		if releaseName == "" && service != nil {
			releaseName = service.Name
		}
		if releaseName != "" {
			return fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName)
		}
	}
	if service != nil {
		return fmt.Sprintf("app=%s", service.Name)
	}
	return ""
}

// 启动前失败的等待原因
var waitingFailReasons = map[string]bool{
	"InvalidImageName":          true,
	"ErrImagePull":              true,
	"ImagePullBackOff":          true,
	"CrashLoopBackOff":          true,
	"CreateContainerConfigError": true,
}

// podHealthCheck 在滚动更新完成后进入 60 秒观察期，检测 Pod 异常
func (w *RolloutWatcher) podHealthCheck(ctx context.Context, clientset *kubernetes.Clientset, deployment *model.Deployment, service *model.Service, room string) {
	ticker := time.NewTicker(podCheckInterval)
	defer ticker.Stop()
	deadline := time.After(podCheckDuration)

	// 记录初始 RestartCount 用于检测观察期内重启
	initialRestarts := map[string]int32{}

	checkCount := 0
	totalChecks := int(podCheckDuration / podCheckInterval)

	for {
		select {
		case <-ctx.Done():
			w.markPodUnhealthy(deployment.ID, "健康检查超时（父上下文取消）")
			w.broadcastProgress(room, deployment.ID, model.DeployStatusPodUnhealthy, 0, 0, "健康检查超时")
			return

		case <-deadline:
			// 观察期结束，无异常，标记为 pod_healthy
			w.markPodHealthy(deployment.ID)
			w.broadcastProgress(room, deployment.ID, model.DeployStatusPodHealthy, 0, 0,
				"Pod 健康检查通过，所有 Pod 在观察期内稳定运行")
			return

		case <-ticker.C:
			checkCount++
			labelSelector := podLabelSelector(deployment, service)
			pods, err := clientset.CoreV1().Pods(deployment.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				log.Printf("[PodHealthCheck] 获取 Pod 列表失败, deploymentID=%d: %v", deployment.ID, err)
				w.broadcastProgress(room, deployment.ID, model.DeployStatusPodChecking, 0, 0,
					fmt.Sprintf("Pod 检查中 (%d/%d)... 获取 Pod 列表失败", checkCount, totalChecks))
				continue
			}

			for _, pod := range pods.Items {
				podKey := pod.Name

				// 检查所有容器状态
				for _, cStatus := range pod.Status.ContainerStatuses {
					// 启动前失败（Waiting 状态）
					if cStatus.State.Waiting != nil {
						if waitingFailReasons[cStatus.State.Waiting.Reason] {
							msg := fmt.Sprintf("Pod %s 容器 %s 启动失败: %s — %s",
								pod.Name, cStatus.Name, cStatus.State.Waiting.Reason, cStatus.State.Waiting.Message)
							w.markPodUnhealthy(deployment.ID, msg)
							w.broadcastProgress(room, deployment.ID, model.DeployStatusPodUnhealthy, 0, 0, msg)
							return
						}
					}

					// 启动后失败（Terminated 状态）
					if cStatus.State.Terminated != nil {
						reason := cStatus.State.Terminated.Reason
						exitCode := cStatus.State.Terminated.ExitCode
						if reason == "OOMKilled" || exitCode != 0 {
							logTail := getPodLogTailFromClientset(ctx, clientset, pod.Name, deployment.Namespace, cStatus.Name)
							msg := fmt.Sprintf("Pod %s 容器 %s 异常退出: %s (exit code: %d)",
								pod.Name, cStatus.Name, reason, exitCode)
							if logTail != "" {
								msg += "\n--- 最近日志 ---\n" + logTail
							}
							w.markPodUnhealthy(deployment.ID, msg)
							w.broadcastProgress(room, deployment.ID, model.DeployStatusPodUnhealthy, 0, 0,
								fmt.Sprintf("Pod %s 容器 %s 异常退出: %s", pod.Name, cStatus.Name, reason))
							return
						}
					}

					// 重启检测
					restartKey := podKey + "/" + cStatus.Name
					if _, exists := initialRestarts[restartKey]; !exists {
						initialRestarts[restartKey] = cStatus.RestartCount
					} else if cStatus.RestartCount > initialRestarts[restartKey] {
						logTail := getPodLogTailFromClientset(ctx, clientset, pod.Name, deployment.Namespace, cStatus.Name)
						msg := fmt.Sprintf("Pod %s 容器 %s 在观察期内发生重启 (重启次数: %d → %d)",
							pod.Name, cStatus.Name, initialRestarts[restartKey], cStatus.RestartCount)
						if cStatus.LastTerminationState.Terminated != nil {
							msg += fmt.Sprintf(", 上次退出原因: %s (exit code: %d)",
								cStatus.LastTerminationState.Terminated.Reason,
								cStatus.LastTerminationState.Terminated.ExitCode)
						}
						if logTail != "" {
							msg += "\n--- 最近日志 ---\n" + logTail
						}
						w.markPodUnhealthy(deployment.ID, msg)
						w.broadcastProgress(room, deployment.ID, model.DeployStatusPodUnhealthy, 0, 0,
							fmt.Sprintf("Pod %s 容器 %s 在观察期内重启", pod.Name, cStatus.Name))
						return
					}
				}
			}

			// 广播进度
			remaining := podCheckDuration - time.Duration(checkCount)*podCheckInterval
			w.broadcastProgress(room, deployment.ID, model.DeployStatusPodChecking, 0, 0,
				fmt.Sprintf("Pod 健康检查中 (%d/%d)... 剩余 %ds", checkCount, totalChecks, int(remaining.Seconds())))
		}
	}
}

// getPodLogTailFromClientset 获取 Pod 容器最近 50 行日志，权限不足时优雅降级返回空
func getPodLogTailFromClientset(ctx context.Context, cs *kubernetes.Clientset, podName, namespace, container string) string {
	tailLines := int64(50)
	req := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return ""
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		return ""
	}
	result := string(data)
	if len(result) > 4096 {
		result = result[len(result)-4096:]
	}
	return strings.TrimSpace(result)
}

func (w *RolloutWatcher) markPodChecking(deploymentID uint) {
	if err := w.deploySvc.UpdatePodStatus(deploymentID, model.DeployStatusPodChecking, "checking", ""); err != nil {
		log.Printf("[RolloutWatcher] 更新为 pod_checking 失败, deploymentID=%d: %v", deploymentID, err)
	}
}

func (w *RolloutWatcher) markPodHealthy(deploymentID uint) {
	now := time.Now()
	if err := w.deploySvc.UpdatePodStatus(deploymentID, model.DeployStatusPodHealthy, "healthy", ""); err != nil {
		log.Printf("[RolloutWatcher] 更新为 pod_healthy 失败, deploymentID=%d: %v", deploymentID, err)
	}
	if dep, err := w.deploySvc.GetByID(deploymentID); err == nil {
		dep.FinishedAt = &now
		_ = w.deploySvc.deployRepo.Update(dep)
		w.dispatchDeployEvent(dep, model.EventDeploySuccess, "")
	}
}

func (w *RolloutWatcher) markPodUnhealthy(deploymentID uint, message string) {
	now := time.Now()
	if err := w.deploySvc.UpdatePodStatus(deploymentID, model.DeployStatusPodUnhealthy, "unhealthy", message); err != nil {
		log.Printf("[RolloutWatcher] 更新为 pod_unhealthy 失败, deploymentID=%d: %v", deploymentID, err)
	}
	if dep, err := w.deploySvc.GetByID(deploymentID); err == nil {
		dep.FinishedAt = &now
		_ = w.deploySvc.deployRepo.Update(dep)
		w.dispatchDeployEvent(dep, model.EventPodUnhealthy, message)
	}
}

func (w *RolloutWatcher) markSuccess(deploymentID uint) {
	if err := w.deploySvc.UpdateStatus(deploymentID, model.DeployStatusSuccess); err != nil {
		log.Printf("[RolloutWatcher] 更新部署状态为成功失败, deploymentID=%d: %v", deploymentID, err)
	}
	if dep, err := w.deploySvc.GetByID(deploymentID); err == nil {
		w.dispatchDeployEvent(dep, model.EventDeploySuccess, "")
	}
}

func (w *RolloutWatcher) markFailed(deploymentID uint, reason string) {
	log.Printf("[RolloutWatcher] 部署失败, deploymentID=%d: %s", deploymentID, reason)
	if err := w.deploySvc.UpdateStatusWithReason(deploymentID, model.DeployStatusFailed, reason); err != nil {
		log.Printf("[RolloutWatcher] 更新部署状态为失败失败, deploymentID=%d: %v", deploymentID, err)
	}
	if dep, err := w.deploySvc.GetByID(deploymentID); err == nil {
		w.dispatchDeployEvent(dep, model.EventDeployFailed, reason)
	}
}

// dispatchDeployEvent 发送部署相关通知
func (w *RolloutWatcher) dispatchDeployEvent(dep *model.Deployment, eventType, failReason string) {
	if w.notifDispatcher == nil {
		return
	}
	svcName := ""
	if svc, err := w.deploySvc.GetServiceByID(dep.ServiceID); err == nil {
		svcName = svc.Name
	}
	w.notifDispatcher.Dispatch(dep.ServiceID, eventType, notification.NotificationPayload{
		ServiceName: svcName,
		Namespace:   dep.Namespace,
		ImageTag:    dep.ImageTag,
		FailReason:  failReason,
		DeployID:    dep.ID,
	})
}

func (w *RolloutWatcher) broadcastProgress(room string, deploymentID uint, status string, replicas, readyReplicas int32, message string) {
	if w.broadcaster == nil {
		return
	}
	progress := RolloutProgress{
		DeploymentID:  deploymentID,
		Status:        status,
		Replicas:      replicas,
		ReadyReplicas: readyReplicas,
		Message:       message,
	}
	data, err := json.Marshal(progress)
	if err != nil {
		log.Printf("[RolloutWatcher] JSON 序列化失败: %v", err)
		return
	}
	w.broadcaster.Broadcast(room, data)
}

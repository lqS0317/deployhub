package build

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"deployhub/internal/ws"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StreamBuildLogs 从 Kaniko Pod 流式读取日志并推送到 WebSocket
func StreamBuildLogs(clientset kubernetes.Interface, namespace, jobName string, hub *ws.Hub, buildID uint, buildSvc *BuildService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 等待 Pod 启动
	podName, err := waitForPod(ctx, clientset, namespace, jobName)
	if err != nil {
		return fmt.Errorf("等待构建 Pod 启动失败: %w", err)
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "kaniko",
		Follow:    true,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("获取 Pod 日志流失败: %w", err)
	}
	defer stream.Close()

	room := fmt.Sprintf("build:%d", buildID)
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text() + "\n"

		if err := buildSvc.AppendLog(buildID, line); err != nil {
			log.Printf("[Build %d] 追加日志失败: %v", buildID, err)
		}

		msg := wsMessage{Type: "log", BuildID: buildID, Log: line}
		data, _ := json.Marshal(msg)
		hub.Broadcast(room, data)
	}

	return scanner.Err()
}

// waitForPod 轮询等待 Job 关联的 Pod 进入 Running 状态
func waitForPod(ctx context.Context, clientset kubernetes.Interface, namespace, jobName string) (string, error) {
	labelSelector := fmt.Sprintf("job-name=%s", jobName)

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("等待 Pod 超时")
		default:
		}

		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded {
				return pod.Name, nil
			}
			// 初始化容器运行完、主容器已启动也可以开始读日志
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == "kaniko" && (cs.State.Running != nil || cs.State.Terminated != nil) {
					return pod.Name, nil
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
}

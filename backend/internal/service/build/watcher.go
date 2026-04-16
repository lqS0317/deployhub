package build

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/service/cluster"
	"deployhub/internal/ws"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// JobWatcher 监控 Kaniko 构建 Job 的状态变化
type JobWatcher struct {
	clientPool *cluster.ClientsetPool
	buildSvc   *BuildService
	hub        *ws.Hub
}

// NewJobWatcher 创建 Job 监视器
func NewJobWatcher(clientPool *cluster.ClientsetPool, buildSvc *BuildService, hub *ws.Hub) *JobWatcher {
	return &JobWatcher{
		clientPool: clientPool,
		buildSvc:   buildSvc,
		hub:        hub,
	}
}

// wsMessage WebSocket 推送的消息结构
type wsMessage struct {
	Type    string `json:"type"`
	BuildID uint   `json:"build_id"`
	Status  string `json:"status,omitempty"`
	Log     string `json:"log,omitempty"`
}

// WatchBuild 使用 client-go Watch API 监控指定构建的 Job 状态
func (w *JobWatcher) WatchBuild(build *model.Build) error {
	clientset, err := w.clientPool.GetClientset(build.BuildClusterID)
	if err != nil {
		return fmt.Errorf("获取集群 clientset 失败: %w", err)
	}

	namespace := "deployhub-builds"
	if err := w.buildSvc.UpdateStatus(build.ID, model.BuildStatusBuilding); err != nil {
		return fmt.Errorf("更新构建状态失败: %w", err)
	}

	w.broadcastStatus(build.ID, model.BuildStatusBuilding)

	go func() {
		defer func() {
			_ = CleanupBuildResources(clientset, namespace, build.KanikoJobName)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		watcher, err := clientset.BatchV1().Jobs(namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", build.KanikoJobName),
		})
		if err != nil {
			log.Printf("[Build %d] 启动 Job Watch 失败: %v", build.ID, err)
			_ = w.buildSvc.UpdateStatus(build.ID, model.BuildStatusFailed)
			w.broadcastStatus(build.ID, model.BuildStatusFailed)
			return
		}
		defer watcher.Stop()

		// 启动日志流式传输
		go func() {
			if err := StreamBuildLogs(clientset, namespace, build.KanikoJobName, w.hub, build.ID, w.buildSvc); err != nil {
				log.Printf("[Build %d] 日志流传输失败: %v", build.ID, err)
			}
		}()

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				log.Printf("[Build %d] Watch 事件错误", build.ID)
				continue
			}

			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			if isJobComplete(job) {
				now := time.Now()
				status := model.BuildStatusSuccess
				if isJobFailed(job) {
					status = model.BuildStatusFailed
				}

				_ = w.buildSvc.UpdateStatus(build.ID, status)
				_ = w.buildSvc.buildRepo.Update(&model.Build{
					ID:         build.ID,
					FinishedAt: &now,
				})
				w.broadcastStatus(build.ID, status)
				return
			}
		}
	}()

	return nil
}

// broadcastStatus 通过 WebSocket 广播构建状态变更
func (w *JobWatcher) broadcastStatus(buildID uint, status string) {
	msg := wsMessage{Type: "status", BuildID: buildID, Status: status}
	data, _ := json.Marshal(msg)
	room := fmt.Sprintf("build:%d", buildID)
	w.hub.Broadcast(room, data)
}

func isJobComplete(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == "True" {
			return true
		}
	}
	return false
}

func isJobFailed(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == "True" {
			return true
		}
	}
	return false
}

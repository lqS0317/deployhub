package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"deployhub/internal/repository"
	"deployhub/internal/service/auth"
	"deployhub/internal/service/cluster"
	"deployhub/internal/ws"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应做域名白名单校验
		return true
	},
}

// HandleBuildLogWS 处理构建日志的 WebSocket 升级请求
// 通过查询参数 ?token=xxx 传递 JWT 进行认证
func HandleBuildLogWS(hub *ws.Hub, jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			return
		}

		if _, err := jwtSvc.ValidateToken(token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			return
		}

		buildID := c.Param("id")
		if buildID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少构建 ID"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		room := "build:" + buildID
		client := ws.NewClient(hub, conn, room)
		hub.Register(room, client)

		go client.WritePump()
		go client.ReadPump()
	}
}

// HandlePodLogWS 处理 Pod 容器日志的 WebSocket 流式推送
// 通过查询参数 ?token=xxx&pod=xxx&container=xxx&tail=100
func HandlePodLogWS(clientPool *cluster.ClientsetPool, deployRepo repository.DeploymentRepository, jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			return
		}
		if _, err := jwtSvc.ValidateToken(token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			return
		}

		deployID := c.Param("id")
		podName := c.Query("pod")
		container := c.Query("container")
		tailStr := c.DefaultQuery("tail", "100")
		tailLines, _ := strconv.ParseInt(tailStr, 10, 64)
		if tailLines <= 0 {
			tailLines = 100
		}

		if podName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 pod 参数"})
			return
		}

		// 获取部署记录
		id, _ := strconv.ParseUint(deployID, 10, 32)
		dep, err := deployRepo.FindByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "部署记录不存在"})
			return
		}

		clientset, err := clientPool.GetClientset(dep.ClusterID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "连接集群失败"})
			return
		}

		ctx := context.Background()
		namespace := dep.Namespace

		// 校验 Pod 状态
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Pod 不存在: %v", err)})
			return
		}

		// 检查容器状态
		if container == "" && len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}

		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == container && cs.State.Waiting != nil {
				c.JSON(http.StatusConflict, gin.H{
					"error":  fmt.Sprintf("容器未就绪: %s", cs.State.Waiting.Reason),
					"reason": cs.State.Waiting.Reason,
					"message": cs.State.Waiting.Message,
				})
				return
			}
		}

		if pod.Status.Phase == corev1.PodPending {
			c.JSON(http.StatusConflict, gin.H{"error": "Pod 未就绪（Pending）"})
			return
		}

		// 升级 WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}
		defer conn.Close()

		// 通过 client-go 流式读取日志
		logOptions := &corev1.PodLogOptions{
			Container: container,
			Follow:    true,
			TailLines: &tailLines,
		}

		stream, err := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions).Stream(ctx)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[ERROR] 获取日志流失败: %v", err)))
			return
		}
		defer stream.Close()

		buf := make([]byte, 4096)
		for {
			n, err := stream.Read(buf)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.TextMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\n[ERROR] 日志流中断: %v", err)))
				}
				return
			}
		}
	}
}

// HandleDeployPodList 获取部署相关的 Pod 列表（HTTP）
func HandleDeployPodList(clientPool *cluster.ClientsetPool, deployRepo repository.DeploymentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		dep, err := deployRepo.FindByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "部署记录不存在"})
			return
		}

		clientset, err := clientPool.GetClientset(dep.ClusterID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "连接集群失败"})
			return
		}

		serviceName := ""
		if dep.Service != nil {
			serviceName = dep.Service.Name
		}

		// Helm 部署使用 app.kubernetes.io/instance 标签，Direct 部署使用 app 标签
		// 先尝试 Helm 标签，如果没找到再尝试 app 标签
		releaseName := dep.HelmReleaseName
		if releaseName == "" && dep.Service != nil {
			releaseName = dep.Service.HelmReleaseName
		}
		if releaseName == "" {
			releaseName = serviceName
		}

		ctx := context.Background()
		var pods *corev1.PodList

		if dep.DeployType == "helm" && releaseName != "" {
			// Helm 部署: 用标准 Helm 标签
			pods, err = clientset.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
			})
			// 如果 Helm 标签没找到 Pod，回退到 app 标签
			if err == nil && len(pods.Items) == 0 && serviceName != "" {
				pods, err = clientset.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("app=%s", serviceName),
				})
			}
		} else if serviceName != "" {
			pods, err = clientset.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", serviceName),
			})
		} else {
			pods, err = clientset.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{})
		}
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("获取 Pod 列表失败: %v", err)})
			return
		}

		type podInfo struct {
			Name       string   `json:"name"`
			Status     string   `json:"status"`
			Ready      bool     `json:"ready"`
			Containers []string `json:"containers"`
			CreatedAt  string   `json:"created_at"`
		}

		var items []podInfo
		for _, p := range pods.Items {
			ready := true
			for _, cs := range p.Status.ContainerStatuses {
				if !cs.Ready {
					ready = false
					break
				}
			}
			var containers []string
			for _, c := range p.Spec.Containers {
				containers = append(containers, c.Name)
			}
			items = append(items, podInfo{
				Name:       p.Name,
				Status:     string(p.Status.Phase),
				Ready:      ready,
				Containers: containers,
				CreatedAt:  p.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
			})
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

// HandleDeployProgressWS 处理部署进度的 WebSocket 升级请求
// 通过查询参数 ?token=xxx 传递 JWT 进行认证，连接到 room deployment:{id}
func HandleDeployProgressWS(hub *ws.Hub, jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			return
		}

		if _, err := jwtSvc.ValidateToken(token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			return
		}

		deployID := c.Param("id")
		if deployID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少部署 ID"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		room := "deployment:" + deployID
		client := ws.NewClient(hub, conn, room)
		hub.Register(room, client)

		go client.WritePump()
		go client.ReadPump()
	}
}

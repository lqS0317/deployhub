package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"
	"deployhub/internal/service/approval"
	"deployhub/internal/service/deploy"
	"deployhub/internal/service/notification"
	"deployhub/internal/service/svc"

	"github.com/gin-gonic/gin"
)

// DeployHandler 部署处理器
type DeployHandler struct {
	deploySvc        *deploy.DeployService
	directExecutor   *deploy.DirectExecutor
	helmExecutor     *deploy.HelmExecutor
	watcher          *deploy.RolloutWatcher
	effectiveRoleSvc *svc.EffectiveRoleService
	approvalSvc      *approval.ApprovalService
	userRepo         repository.UserRepository
	notifDispatcher  *notification.Dispatcher
}

// NewDeployHandler 创建部署处理器
func NewDeployHandler(deploySvc *deploy.DeployService, directExecutor *deploy.DirectExecutor, helmExecutor *deploy.HelmExecutor, watcher *deploy.RolloutWatcher, effectiveRoleSvc *svc.EffectiveRoleService, approvalSvc *approval.ApprovalService, userRepo repository.UserRepository, notifDispatcher *notification.Dispatcher) *DeployHandler {
	return &DeployHandler{
		deploySvc:        deploySvc,
		directExecutor:   directExecutor,
		helmExecutor:     helmExecutor,
		watcher:          watcher,
		effectiveRoleSvc: effectiveRoleSvc,
		approvalSvc:      approvalSvc,
		userRepo:         userRepo,
		notifDispatcher:  notifDispatcher,
	}
}

type createDeploymentRequest struct {
	ServiceID       uint   `json:"service_id" binding:"required"`
	ClusterID       uint   `json:"cluster_id"`
	Namespace       string `json:"namespace"`
	BuildID         *uint  `json:"build_id"`
	ImageTag        string `json:"image_tag"`
	ImageSource     string `json:"image_source"`
	ExternalImage   string `json:"external_image"`
	DeployType      string `json:"deploy_type"`
	WorkloadType    string `json:"workload_type"`
	HealthCheckPath string `json:"health_check_path"`
	HelmRepoID      *uint  `json:"helm_repo_id"`
	HelmChartPath   string `json:"helm_chart_path"`
	HelmReleaseName string `json:"helm_release_name"`
	HelmChartBranch    string `json:"helm_chart_branch"`
	HelmServiceAccount string `json:"helm_service_account"`
}

// Create 创建部署
func (h *DeployHandler) Create(c *gin.Context) {
	var req createDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)
	imageSource := req.ImageSource
	if imageSource == "" {
		imageSource = "build"
	}
	deployType := req.DeployType
	if deployType == "" {
		deployType = "direct"
	}
	workloadType := req.WorkloadType
	if workloadType == "" {
		workloadType = "deployment"
	}

	dep, err := h.deploySvc.CreateDeployment(req.ServiceID, userID, req.ClusterID, req.Namespace, req.BuildID, req.ImageTag, imageSource, req.ExternalImage, deploy.DeployConfig{
		DeployType:         deployType,
		WorkloadType:       workloadType,
		HealthCheckPath:    req.HealthCheckPath,
		HelmRepoID:         req.HelmRepoID,
		HelmChartPath:      req.HelmChartPath,
		HelmReleaseName:    req.HelmReleaseName,
		HelmChartBranch:    req.HelmChartBranch,
		HelmServiceAccount: req.HelmServiceAccount,
	})
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	// 创建后自动触发预览
	go h.autoPreview(dep.ID, dep.ServiceID, userID)

	c.JSON(http.StatusCreated, dep)
}

// autoPreview 创建部署后自动触发预览，完成后根据角色决定审批/放行
func (h *DeployHandler) autoPreview(deploymentID, serviceID, triggerUserID uint) {
	dep, err := h.deploySvc.GetByID(deploymentID)
	if err != nil {
		log.Printf("[AutoPreview] 部署 %d 不存在: %v", deploymentID, err)
		return
	}

	svcModel, err := h.deploySvc.GetServiceByID(serviceID)
	if err != nil {
		log.Printf("[AutoPreview] 服务 %d 不存在: %v", serviceID, err)
		h.deploySvc.UpdateStatusWithReason(deploymentID, model.DeployStatusFailed, fmt.Sprintf("服务不存在: %v", err))
		return
	}

	deployType := dep.DeployType
	if deployType == "" {
		deployType = svcModel.DeployType
	}

	var previewYAML string
	var summaryJSON []byte

	if deployType == "helm" && h.helmExecutor != nil {
		yamlOutput, helmErr := h.helmExecutor.Preview(dep, svcModel)
		if helmErr != nil {
			log.Printf("[AutoPreview] Helm 预览失败: %v", helmErr)
			h.deploySvc.UpdateStatusWithReason(deploymentID, model.DeployStatusFailed, fmt.Sprintf("Helm 预览失败: %v", helmErr))
			return
		}
		previewYAML = deploy.SanitizeSecrets(yamlOutput)
		summaryJSON = []byte(`{"resources":[],"source":"helm_template"}`)
	} else if h.directExecutor != nil {
		yaml, summary, dryErr := h.directExecutor.DryRun(dep, svcModel)
		if dryErr != nil {
			log.Printf("[AutoPreview] Direct dry-run 失败: %v", dryErr)
			h.deploySvc.UpdateStatusWithReason(deploymentID, model.DeployStatusFailed, fmt.Sprintf("Direct dry-run 失败: %v", dryErr))
			return
		}
		previewYAML = yaml
		summaryJSON = summary
	} else {
		previewYAML = "# 执行器未配置"
		summaryJSON = []byte(`{"resources":[]}`)
	}

	// 保存预览结果（状态 → previewed）
	_ = h.deploySvc.SavePreview(deploymentID, previewYAML, summaryJSON)

	// 根据发起人角色决定审批/放行
	if h.userRepo == nil {
		// 测试环境没有 userRepo，直接放行
		h.deploySvc.UpdateStatus(deploymentID, model.DeployStatusApproved)
		return
	}
	user, err := h.userRepo.FindByID(triggerUserID)
	if err != nil {
		log.Printf("[AutoPreview] 查询用户失败: %v", err)
		return
	}

	if user.Role == "admin" {
		// Admin 自动放行
		h.deploySvc.UpdateStatus(deploymentID, model.DeployStatusApproved)
		log.Printf("[AutoPreview] 部署 %d: Admin 自动放行", deploymentID)
	} else {
		// 非 Admin: 创建审批记录
		if h.approvalSvc != nil {
			if err := h.approvalSvc.CreateApprovalForAdmins(deploymentID, triggerUserID); err != nil {
				log.Printf("[AutoPreview] 创建审批失败: %v, 自动放行", err)
				h.deploySvc.UpdateStatus(deploymentID, model.DeployStatusApproved)
			} else {
				log.Printf("[AutoPreview] 部署 %d: 已创建审批记录，等待 Admin 审批", deploymentID)
			}
		} else {
			h.deploySvc.UpdateStatus(deploymentID, model.DeployStatusApproved)
		}
	}
}

// List 部署列表
func (h *DeployHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)

	var serviceID *uint
	if sid := c.Query("service_id"); sid != "" {
		id, err := strconv.ParseUint(sid, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 service_id")
			return
		}
		uid := uint(id)
		serviceID = &uid
	}

	items, total, err := h.deploySvc.List(page, pageSize, serviceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询部署列表失败")
		return
	}

	userRole := middleware.GetUserRole(c)
	if userRole != "admin" && h.effectiveRoleSvc != nil {
		userID := middleware.GetUserID(c)
		var filtered []model.Deployment
		for _, d := range items {
			role, _ := h.effectiveRoleSvc.GetEffectiveRole(userID, d.ServiceID)
			if role != "" {
				filtered = append(filtered, d)
			}
		}
		pkg.Paginated(c, filtered, int64(len(filtered)), page, pageSize)
		return
	}

	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 部署详情
func (h *DeployHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	c.JSON(http.StatusOK, dep)
}

// Rollback 触发回滚
func (h *DeployHandler) Rollback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	userID := middleware.GetUserID(c)
	dep, err := h.deploySvc.Rollback(uint(id), userID)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	// 发送回滚触发通知
	if h.notifDispatcher != nil {
		payload := notification.NotificationPayload{
			Namespace: dep.Namespace,
			ImageTag:  dep.ImageTag,
			DeployID:  dep.ID,
		}
		if svcModel, err := h.deploySvc.GetServiceByID(dep.ServiceID); err == nil {
			payload.ServiceName = svcModel.Name
		}
		if user, err := h.userRepo.FindByID(userID); err == nil {
			payload.TriggerUser = user.Username
		}
		h.notifDispatcher.Dispatch(dep.ServiceID, model.EventRollbackTriggered, payload)
	}

	c.JSON(http.StatusCreated, dep)
}

// ExecuteDeploy 触发部署执行（审批通过后调用）
func (h *DeployHandler) ExecuteDeploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	// 允许 pending_approval 或 approved 状态触发执行
	if dep.Status != model.DeployStatusPendingApproval && dep.Status != model.DeployStatusApproved {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, "当前状态无法执行部署")
		return
	}

	if err := h.deploySvc.StartDeploy(uint(id)); err != nil {
		// StartDeploy 要求 approved 状态，如果是 pending_approval 先改状态
		h.deploySvc.UpdateStatus(uint(id), model.DeployStatusApproved)
		if err := h.deploySvc.StartDeploy(uint(id)); err != nil {
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
			return
		}
	}

	// 重新加载含 Service 信息
	dep, _ = h.deploySvc.GetByID(uint(id))
	svc, err := h.deploySvc.GetServiceByID(dep.ServiceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "服务不存在")
		return
	}

	// 根据 deploy_type 路由到对应执行器（优先 deployment，fallback service）
	execDeployType := dep.DeployType
	if execDeployType == "" {
		execDeployType = svc.DeployType
	}
	if execDeployType == "helm" {
		if h.helmExecutor != nil {
			if err := h.helmExecutor.Execute(dep, svc); err != nil {
				h.deploySvc.UpdateStatusWithReason(uint(id), model.DeployStatusFailed, err.Error())
				pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Helm 部署已触发"})
			return
		}
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "Helm 执行器未配置")
		return
	}

	// Direct 部署
	if h.directExecutor != nil {
		if err := h.directExecutor.Execute(dep, svc); err != nil {
			h.deploySvc.UpdateStatusWithReason(uint(id), model.DeployStatusFailed, err.Error())
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
			return
		}
	}
	if h.watcher != nil {
		h.watcher.Watch(dep, svc)
	}

	c.JSON(http.StatusOK, gin.H{"message": "部署已触发"})
}

// Preview 触发 dry-run 预览
func (h *DeployHandler) Preview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	if err := h.deploySvc.StartPreview(uint(id)); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	svc, err := h.deploySvc.GetServiceByID(dep.ServiceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "服务不存在")
		return
	}

	var previewYAML string
	var summaryJSON []byte

	// 优先从 deployment 读取 deploy_type，fallback 到 service
	deployType := dep.DeployType
	if deployType == "" {
		deployType = svc.DeployType
	}

	if deployType == "helm" && h.helmExecutor != nil {
		// Helm dry-run: 异步执行 helm template
		helmExec := h.helmExecutor
		deploySvc := h.deploySvc
		go func() {
			yamlOutput, helmErr := helmExec.Preview(dep, svc)
			if helmErr != nil {
				deploySvc.UpdateStatusWithReason(uint(id), model.DeployStatusFailed, fmt.Sprintf("Helm 预览失败: %v", helmErr))
				return
			}
			previewYAML := deploy.SanitizeSecrets(yamlOutput)
			summaryJSON := []byte(`{"resources":[],"source":"helm_template"}`)
			_ = deploySvc.SavePreview(uint(id), previewYAML, summaryJSON)
		}()
		c.JSON(http.StatusOK, gin.H{"message": "Helm 预览已触发，请稍后查看结果"})
		return
	} else if h.directExecutor != nil {
		// Direct dry-run: K8s server-side dry-run（同步，通常很快）
		yaml, summary, dryErr := h.directExecutor.DryRun(dep, svc)
		if dryErr != nil {
			h.deploySvc.UpdateStatusWithReason(uint(id), model.DeployStatusFailed, fmt.Sprintf("dry-run 失败: %v", dryErr))
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, dryErr.Error())
			return
		}
		previewYAML = yaml
		summaryJSON = summary
	} else {
		previewYAML = "# 执行器未配置"
		summaryJSON = []byte(`{"resources":[]}`)
	}

	_ = h.deploySvc.SavePreview(uint(id), previewYAML, summaryJSON)

	c.JSON(http.StatusOK, gin.H{"message": "预览完成", "preview_yaml": previewYAML})
}

// GetPreview 获取预览结果
func (h *DeployHandler) GetPreview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preview_yaml":    dep.PreviewYAML,
		"preview_summary": dep.PreviewSummary,
		"deploy_command":  dep.DeployCommand,
		"status":          dep.Status,
	})
}

// Cancel 取消部署
func (h *DeployHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	// 在取消前获取部署信息用于通知
	dep, _ := h.deploySvc.GetByID(uint(id))

	if err := h.deploySvc.CancelDeploy(uint(id)); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	// 发送部署取消通知
	if h.notifDispatcher != nil && dep != nil {
		userID := middleware.GetUserID(c)
		payload := notification.NotificationPayload{
			Namespace: dep.Namespace,
			ImageTag:  dep.ImageTag,
			DeployID:  dep.ID,
		}
		if svcModel, err := h.deploySvc.GetServiceByID(dep.ServiceID); err == nil {
			payload.ServiceName = svcModel.Name
		}
		if user, err := h.userRepo.FindByID(userID); err == nil {
			payload.TriggerUser = user.Username
		}
		h.notifDispatcher.Dispatch(dep.ServiceID, model.EventDeployCancelled, payload)
	}

	c.JSON(http.StatusOK, gin.H{"message": "部署已取消"})
}

type helmRollbackRequest struct {
	Revision int `json:"revision" binding:"required,min=1"`
}

// HelmRollback 通过 Helm 原生回滚
func (h *DeployHandler) HelmRollback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	var req helmRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "请指定回滚的 revision")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	svc, err := h.deploySvc.GetServiceByID(dep.ServiceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "服务不存在")
		return
	}

	if svc.DeployType != "helm" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "仅 Helm 类型服务支持 revision 回滚")
		return
	}

	if h.helmExecutor == nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "Helm 执行器未配置")
		return
	}

	if err := h.helmExecutor.Rollback(dep.ID, svc, dep.Namespace, dep.ClusterID, req.Revision); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Helm 回滚已触发"})
}

// UpdateDeploy 编辑部署（仅 pending_approval 状态）
func (h *DeployHandler) UpdateDeploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	if dep.Status != model.DeployStatusPendingApproval && dep.Status != model.DeployStatusApproved {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, "仅待审批/已审批状态的部署可编辑")
		return
	}

	var req createDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	// 更新可编辑字段
	if req.ClusterID > 0 {
		dep.ClusterID = req.ClusterID
	}
	if req.Namespace != "" {
		dep.Namespace = req.Namespace
	}
	if req.ImageTag != "" {
		dep.ImageTag = req.ImageTag
	}
	if req.ExternalImage != "" {
		dep.ExternalImage = req.ExternalImage
	}
	if req.DeployType != "" {
		dep.DeployType = req.DeployType
	}
	if req.WorkloadType != "" {
		dep.WorkloadType = req.WorkloadType
	}
	if req.HelmRepoID != nil {
		dep.HelmRepoID = req.HelmRepoID
	}
	if req.HelmChartPath != "" {
		dep.HelmChartPath = req.HelmChartPath
	}
	if req.HelmReleaseName != "" {
		dep.HelmReleaseName = req.HelmReleaseName
	}
	if req.HelmChartBranch != "" {
		dep.HelmChartBranch = req.HelmChartBranch
	}
	if req.HelmServiceAccount != "" {
		dep.HelmServiceAccount = req.HelmServiceAccount
	}

	// 清空预览结果，重置状态，重新触发预览
	dep.PreviewYAML = nil
	dep.PreviewSummary = nil
	dep.Status = model.DeployStatusPreviewing
	dep.FailReason = ""

	if err := h.deploySvc.UpdateDeploy(dep); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	// 自动重新预览
	userID := middleware.GetUserID(c)
	go h.autoPreview(dep.ID, dep.ServiceID, userID)

	c.JSON(http.StatusOK, dep)
}

// DeleteDeploy 删除部署记录
func (h *DeployHandler) DeleteDeploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}
	if err := h.deploySvc.DeleteDeployment(uint(id)); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPods 获取部署相关的 Pod 列表
func (h *DeployHandler) ListPods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的部署 ID")
		return
	}

	dep, err := h.deploySvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "部署记录不存在")
		return
	}

	svcModel, err := h.deploySvc.GetServiceByID(dep.ServiceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "服务不存在")
		return
	}

	// 需要 clientPool — 通过 watcher 或直接注入
	// 由于 handler 没有直接的 clientPool，返回提示
	c.JSON(http.StatusOK, gin.H{
		"items":       []interface{}{},
		"cluster_id":  dep.ClusterID,
		"namespace":   dep.Namespace,
		"service":     svcModel.Name,
		"message":     "Pod 列表需要集群连接，请使用 kubectl 查看",
	})
}

// RegisterDeployRoutes 注册部署相关路由
func RegisterDeployRoutes(r *gin.RouterGroup, h *DeployHandler) {
	deployments := r.Group("/deployments")
	{
		deployments.POST("", h.Create)
		deployments.GET("", h.List)
		deployments.GET("/:id", h.Get)
		deployments.POST("/:id/rollback", h.Rollback)
		deployments.POST("/:id/execute", h.ExecuteDeploy)
		deployments.POST("/:id/preview", h.Preview)
		deployments.GET("/:id/preview", h.GetPreview)
		deployments.POST("/:id/cancel", h.Cancel)
		deployments.POST("/:id/helm-rollback", h.HelmRollback)
		deployments.PUT("/:id", h.UpdateDeploy)
		deployments.DELETE("/:id", h.DeleteDeploy)
		// Pod 列表路由在 main.go 中通过 HandleDeployPodList 注册
	}
}

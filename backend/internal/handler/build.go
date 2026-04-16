package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/service/build"
	"deployhub/internal/service/svc"

	"github.com/gin-gonic/gin"
)

// BuildHandler 构建处理器
type BuildHandler struct {
	buildSvc         *build.BuildService
	buildExecutor    *build.BuildExecutor
	effectiveRoleSvc *svc.EffectiveRoleService
}

// NewBuildHandler 创建构建处理器
func NewBuildHandler(buildSvc *build.BuildService, buildExecutor *build.BuildExecutor, effectiveRoleSvc *svc.EffectiveRoleService) *BuildHandler {
	return &BuildHandler{buildSvc: buildSvc, buildExecutor: buildExecutor, effectiveRoleSvc: effectiveRoleSvc}
}

type triggerBuildRequest struct {
	ServiceID      uint   `json:"service_id" binding:"required"`
	BuildClusterID uint   `json:"build_cluster_id"`
	GitBranch      string `json:"git_branch" binding:"required"`
	GitCommit      string `json:"git_commit"`
	ImageTag       string `json:"image_tag"`
	Name           string `json:"name"`
	DockerfilePath string `json:"dockerfile_path"`
	RegistryID     *uint  `json:"registry_id"`
	ImageRepo      string `json:"image_repo"`
	BuildContext   string `json:"build_context"`
}

// TriggerBuild 触发构建
func (h *BuildHandler) TriggerBuild(c *gin.Context) {
	var req triggerBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		pkg.Error(c, http.StatusUnauthorized, pkg.CodeUnauthorized, "未授权")
		return
	}

	result, err := h.buildSvc.CreateBuild(req.ServiceID, userID, req.BuildClusterID, req.GitBranch, req.GitCommit, req.ImageTag, req.Name, req.DockerfilePath, req.RegistryID, req.ImageRepo, req.BuildContext)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	// 异步触发构建执行
	if h.buildExecutor != nil {
		h.buildExecutor.Execute(result.ID)
	}

	pkg.Success(c, http.StatusCreated, result)
}

// ListBuilds 列出构建记录
func (h *BuildHandler) ListBuilds(c *gin.Context) {
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

	items, total, err := h.buildSvc.List(page, pageSize, serviceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询构建列表失败")
		return
	}

	userRole := middleware.GetUserRole(c)
	if userRole != "admin" && h.effectiveRoleSvc != nil {
		userID := middleware.GetUserID(c)
		var filtered []model.Build
		for _, b := range items {
			role, _ := h.effectiveRoleSvc.GetEffectiveRole(userID, b.ServiceID)
			if role != "" {
				filtered = append(filtered, b)
			}
		}
		pkg.Paginated(c, filtered, int64(len(filtered)), page, pageSize)
		return
	}

	pkg.Paginated(c, items, total, page, pageSize)
}

// GetBuild 获取构建详情
func (h *BuildHandler) GetBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的构建 ID")
		return
	}

	result, err := h.buildSvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "构建记录不存在")
		return
	}

	pkg.Success(c, http.StatusOK, result)
}

// CancelBuild 取消构建
func (h *BuildHandler) CancelBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的构建 ID")
		return
	}

	if err := h.buildSvc.Cancel(uint(id)); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}

	// 删除 K8s Job 并触发通知
	if h.buildExecutor != nil {
		go h.buildExecutor.CancelJob(uint(id))
		h.buildExecutor.DispatchEvent(uint(id), model.EventBuildCancelled, "用户取消构建")
	}

	pkg.Success(c, http.StatusOK, gin.H{"message": "构建已取消"})
}

// GetBuildLog 获取构建日志
func (h *BuildHandler) GetBuildLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的构建 ID")
		return
	}

	log, err := h.buildSvc.GetLog(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "构建记录不存在")
		return
	}

	pkg.Success(c, http.StatusOK, gin.H{"log": log})
}

// RetryBuild 基于已有构建记录重新触发构建
func (h *BuildHandler) RetryBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的构建 ID")
		return
	}

	oldBuild, err := h.buildSvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "构建记录不存在")
		return
	}

	userID := middleware.GetUserID(c)
	result, err := h.buildSvc.CreateBuild(oldBuild.ServiceID, userID, oldBuild.BuildClusterID, oldBuild.GitBranch, oldBuild.GitCommit, "", oldBuild.Name, oldBuild.DockerfilePath, oldBuild.RegistryID, oldBuild.ImageRepo, oldBuild.BuildContext)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	if h.buildExecutor != nil {
		h.buildExecutor.Execute(result.ID)
	}

	pkg.Success(c, http.StatusCreated, result)
}

// DeleteBuild 删除构建记录
func (h *BuildHandler) DeleteBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的构建 ID")
		return
	}
	if err := h.buildSvc.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterBuildRoutes 注册构建相关路由
func RegisterBuildRoutes(r *gin.RouterGroup, h *BuildHandler) {
	builds := r.Group("/builds")
	{
		builds.POST("", h.TriggerBuild)
		builds.GET("", h.ListBuilds)
		builds.GET("/:id", h.GetBuild)
		builds.POST("/:id/cancel", h.CancelBuild)
		builds.POST("/:id/retry", h.RetryBuild)
		builds.DELETE("/:id", h.DeleteBuild)
		builds.GET("/:id/log", h.GetBuildLog)
	}
}

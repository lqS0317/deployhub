package handler

import (
	"errors"
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/service/routing"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// RouteEntryHandler 路由条目处理器
type RouteEntryHandler struct {
	svc *routing.RouteService
}

// NewRouteEntryHandler 创建路由条目处理器
func NewRouteEntryHandler(svc *routing.RouteService) *RouteEntryHandler {
	return &RouteEntryHandler{svc: svc}
}

// ListEntries 列出路由条目
func (h *RouteEntryHandler) ListEntries(c *gin.Context) {
	resourceType := c.Query("resource_type")
	entries, err := h.svc.ListEntries(resourceType)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询路由条目失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": entries})
}

type createRouteEntryRequest struct {
	Name         string         `json:"name" binding:"required"`
	ResourceType string         `json:"resource_type" binding:"required"`
	Config       datatypes.JSON `json:"config" binding:"required"`
}

// CreateEntry 创建路由条目
func (h *RouteEntryHandler) CreateEntry(c *gin.Context) {
	var req createRouteEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	entry := &model.RouteEntry{
		Name:         req.Name,
		ResourceType: req.ResourceType,
		Config:       req.Config,
		CreatedByID:  middleware.GetUserID(c),
	}
	if err := h.svc.CreateEntry(entry); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, entry)
}

// GetEntry 获取单个路由条目
func (h *RouteEntryHandler) GetEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	entry, err := h.svc.GetEntry(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "路由条目不存在")
		return
	}
	c.JSON(http.StatusOK, entry)
}

type updateRouteEntryRequest struct {
	Name         string         `json:"name"`
	ResourceType string         `json:"resource_type"`
	Config       datatypes.JSON `json:"config"`
}

// UpdateEntry 更新路由条目
func (h *RouteEntryHandler) UpdateEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}

	existing, err := h.svc.GetEntry(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "路由条目不存在")
		return
	}

	var req updateRouteEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.ResourceType != "" {
		existing.ResourceType = req.ResourceType
	}
	if req.Config != nil {
		existing.Config = req.Config
	}

	if err := h.svc.UpdateEntry(existing); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteEntry 删除路由条目
func (h *RouteEntryHandler) DeleteEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	if err := h.svc.DeleteEntry(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

type deployRouteRequest struct {
	ClusterID uint   `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// DeployEntry 部署路由条目到集群
func (h *RouteEntryHandler) DeployEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	var req deployRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	if err := h.svc.DeployEntry(uint(id), req.ClusterID, req.Namespace); err != nil {
		if errors.Is(err, routing.ErrNamespaceNotMapped) {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
			return
		}
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "部署成功"})
}

// ListDeployments 列出路由条目的部署记录
func (h *RouteEntryHandler) ListDeployments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	list, err := h.svc.GetDeployments(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询部署记录失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// PreviewEntry 预览路由条目生成的 YAML
func (h *RouteEntryHandler) PreviewEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	namespace := c.DefaultQuery("namespace", "default")
	yamlContent, err := h.svc.PreviewEntry(uint(id), namespace)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"yaml": yamlContent})
}

// RegisterRouteRoutes 注册路由中心路由
func RegisterRouteRoutes(r *gin.RouterGroup, h *RouteEntryHandler, permH *RoutePermissionHandler) {
	routes := r.Group("/route-entries")
	{
		routes.GET("", h.ListEntries)
		routes.POST("", h.CreateEntry)
		routes.GET("/:id", h.GetEntry)
		routes.PUT("/:id", h.UpdateEntry)
		routes.DELETE("/:id", h.DeleteEntry)
		routes.POST("/:id/deploy", h.DeployEntry)
		routes.GET("/:id/deployments", h.ListDeployments)
		routes.GET("/:id/preview", h.PreviewEntry)
	}

	perms := r.Group("/route-permissions")
	{
		perms.GET("", permH.List)
		perms.POST("", permH.Grant)
		perms.DELETE("/:id", permH.Revoke)
	}
}

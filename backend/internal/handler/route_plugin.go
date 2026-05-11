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
)

// RoutePluginHandler 路由插件处理器
type RoutePluginHandler struct {
	svc *routing.PluginService
}

// NewRoutePluginHandler 创建路由插件处理器
func NewRoutePluginHandler(svc *routing.PluginService) *RoutePluginHandler {
	return &RoutePluginHandler{svc: svc}
}

// ListPlugins 列出所有插件
func (h *RoutePluginHandler) ListPlugins(c *gin.Context) {
	list, err := h.svc.ListPlugins()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询插件失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type createPluginRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	YAMLContent string `json:"yaml_content" binding:"required"`
}

// CreatePlugin 创建插件
func (h *RoutePluginHandler) CreatePlugin(c *gin.Context) {
	var req createPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	plugin := &model.RoutePlugin{
		Name:        req.Name,
		Description: req.Description,
		YAMLContent: req.YAMLContent,
		CreatedByID: middleware.GetUserID(c),
	}
	if err := h.svc.CreatePlugin(plugin); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, plugin)
}

// GetPlugin 获取单个插件
func (h *RoutePluginHandler) GetPlugin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	plugin, err := h.svc.GetPlugin(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "插件不存在")
		return
	}
	c.JSON(http.StatusOK, plugin)
}

type updatePluginRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	YAMLContent string `json:"yaml_content"`
}

// UpdatePlugin 更新插件
func (h *RoutePluginHandler) UpdatePlugin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	existing, err := h.svc.GetPlugin(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "插件不存在")
		return
	}

	var req updatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.YAMLContent != "" {
		existing.YAMLContent = req.YAMLContent
	}

	if err := h.svc.UpdatePlugin(existing); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeletePlugin 删除插件
func (h *RoutePluginHandler) DeletePlugin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	if err := h.svc.DeletePlugin(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

type deployPluginRequest struct {
	ClusterID uint   `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// DeployPlugin 部署插件到集群
func (h *RoutePluginHandler) DeployPlugin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	var req deployPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	if err := h.svc.DeployPlugin(uint(id), req.ClusterID, req.Namespace); err != nil {
		if errors.Is(err, routing.ErrNamespaceNotMapped) {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
			return
		}
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "部署成功"})
}

// ListPluginDeployments 列出插件的部署记录
func (h *RoutePluginHandler) ListPluginDeployments(c *gin.Context) {
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

// RegisterPluginRoutes 注册插件中心路由
func RegisterPluginRoutes(r *gin.RouterGroup, h *RoutePluginHandler) {
	plugins := r.Group("/route-plugins")
	{
		plugins.GET("", h.ListPlugins)
		plugins.POST("", h.CreatePlugin)
		plugins.GET("/:id", h.GetPlugin)
		plugins.PUT("/:id", h.UpdatePlugin)
		plugins.DELETE("/:id", h.DeletePlugin)
		plugins.POST("/:id/deploy", h.DeployPlugin)
		plugins.GET("/:id/deployments", h.ListPluginDeployments)
	}
}

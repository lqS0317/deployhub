package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	configSvc "deployhub/internal/service/config"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置中心处理器
type ConfigHandler struct {
	configSvc *configSvc.ConfigService
}

// NewConfigHandler 创建配置中心处理器
func NewConfigHandler(cs *configSvc.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: cs}
}

type createTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	ConfigType      string `json:"config_type" binding:"required"`
	TemplateContent string `json:"template_content" binding:"required"`
}

// ListTemplates 列出服务下的所有配置模板
func (h *ConfigHandler) ListTemplates(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的服务 ID")
		return
	}

	list, err := h.configSvc.ListTemplates(uint(serviceID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询配置模板列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": list})
}

// CreateTemplate 创建配置模板
func (h *ConfigHandler) CreateTemplate(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的服务 ID")
		return
	}

	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	tpl, err := h.configSvc.CreateTemplate(uint(serviceID), req.Name, req.ConfigType, req.TemplateContent)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	c.JSON(http.StatusCreated, tpl)
}

// GetTemplate 获取配置模板详情
func (h *ConfigHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	tpl, err := h.configSvc.GetTemplate(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "配置模板不存在")
		return
	}

	c.JSON(http.StatusOK, tpl)
}

type updateTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	TemplateContent string `json:"template_content" binding:"required"`
}

// UpdateTemplate 更新配置模板
func (h *ConfigHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	tpl, err := h.configSvc.UpdateTemplate(uint(id), req.Name, req.TemplateContent)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tpl)
}

// DeleteTemplate 删除配置模板
func (h *ConfigHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	if err := h.configSvc.DeleteTemplate(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListEnvValues 列出模板下各集群的环境变量摘要
func (h *ConfigHandler) ListEnvValues(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	list, err := h.configSvc.ListEnvValues(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询环境变量失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": list})
}

type setEnvValuesRequest struct {
	Values map[string]string `json:"values" binding:"required"`
}

// SetEnvValues 设置指定集群的环境变量值
func (h *ConfigHandler) SetEnvValues(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("cluster_id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	var req setEnvValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if err := h.configSvc.SetEnvValues(uint(id), uint(clusterID), req.Values); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "环境变量已更新"})
}

type renderRequest struct {
	ClusterID uint `json:"cluster_id" binding:"required"`
}

// RenderPreview 预览渲染结果
func (h *ConfigHandler) RenderPreview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	var req renderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	rendered, err := h.configSvc.RenderPreview(uint(id), req.ClusterID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"rendered": rendered})
}

// ListVersions 列出配置版本列表
func (h *ConfigHandler) ListVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	clusterIDStr := c.Query("cluster_id")
	if clusterIDStr == "" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "缺少 cluster_id 参数")
		return
	}

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	list, err := h.configSvc.ListVersions(uint(id), uint(clusterID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询版本列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": list})
}

// DiffVersions 对比两个版本的配置差异
func (h *ConfigHandler) DiffVersions(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	versionID, err := strconv.ParseUint(c.Param("version_id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的版本 ID")
		return
	}

	compareStr := c.Query("compare_with")
	if compareStr == "" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "缺少 compare_with 参数")
		return
	}

	compareID, err := strconv.ParseUint(compareStr, 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的对比版本 ID")
		return
	}

	oldVer, err := h.configSvc.GetVersion(uint(versionID))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "版本不存在")
		return
	}

	newVer, err := h.configSvc.GetVersion(uint(compareID))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "对比版本不存在")
		return
	}

	diff := configSvc.DiffVersions(oldVer.RenderedContent, newVer.RenderedContent)

	c.JSON(http.StatusOK, gin.H{
		"version_id":   versionID,
		"compare_with": compareID,
		"diff":         diff,
	})
}

type deployRequest struct {
	VersionID uint   `json:"version_id" binding:"required"`
	ClusterID uint   `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// DeployConfig 下发配置到 K8s 集群
func (h *ConfigHandler) DeployConfig(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置模板 ID")
		return
	}

	var req deployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)

	dep, err := h.configSvc.DeployConfig(req.VersionID, req.ClusterID, userID, req.Namespace)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, dep)
}

// RegisterConfigRoutes 注册配置中心路由
func RegisterConfigRoutes(r *gin.RouterGroup, h *ConfigHandler) {
	// 服务级别的配置模板路由
	services := r.Group("/services")
	{
		services.GET("/:id/configs", h.ListTemplates)
		services.POST("/:id/configs", h.CreateTemplate)
	}

	// 配置模板级别的路由
	configs := r.Group("/configs")
	{
		configs.GET("/:id", h.GetTemplate)
		configs.PUT("/:id", h.UpdateTemplate)
		configs.DELETE("/:id", h.DeleteTemplate)
		configs.GET("/:id/env-values", h.ListEnvValues)
		configs.PUT("/:id/env-values/:cluster_id", h.SetEnvValues)
		configs.POST("/:id/render", h.RenderPreview)
		configs.GET("/:id/versions", h.ListVersions)
		configs.GET("/:id/versions/:version_id/diff", h.DiffVersions)
		configs.POST("/:id/deploy", h.DeployConfig)
	}
}

package handler

import (
	"net/http"

	"deployhub/internal/pkg"
	"deployhub/internal/service/setting"

	"github.com/gin-gonic/gin"
)

// SystemSettingHandler 系统配置处理器
type SystemSettingHandler struct {
	settingSvc *setting.SettingService
}

func NewSystemSettingHandler(settingSvc *setting.SettingService) *SystemSettingHandler {
	return &SystemSettingHandler{settingSvc: settingSvc}
}

// List 获取所有系统配置
func (h *SystemSettingHandler) List(c *gin.Context) {
	settings, err := h.settingSvc.GetAll()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "获取系统配置失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": settings})
}

type updateSettingRequest struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

// Update 更新单个配置项
func (h *SystemSettingHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "缺少配置 key")
		return
	}

	var req updateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if err := h.settingSvc.Set(key, req.Value, req.Description); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "更新配置失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": req.Value})
}

// RegisterSystemSettingRoutes 注册系统配置路由
func RegisterSystemSettingRoutes(r *gin.RouterGroup, h *SystemSettingHandler) {
	settings := r.Group("/system-settings")
	{
		settings.GET("", h.List)
		settings.PUT("/:key", h.Update)
	}
}

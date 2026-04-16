package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/pkg"
	"deployhub/internal/service/registry"

	"github.com/gin-gonic/gin"
)

// RegistryHandler 镜像仓库处理器
type RegistryHandler struct {
	svc *registry.RegistryService
}

// NewRegistryHandler 创建镜像仓库处理器
func NewRegistryHandler(svc *registry.RegistryService) *RegistryHandler {
	return &RegistryHandler{svc: svc}
}

type createRegistryRequest struct {
	Name       string `json:"name" binding:"required"`
	URL        string `json:"url" binding:"required"`
	Provider   string `json:"provider" binding:"required"`
	AuthConfig string `json:"auth_config" binding:"required"`
}

// Create 创建镜像仓库
func (h *RegistryHandler) Create(c *gin.Context) {
	var req createRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	result, err := h.svc.Create(req.Name, req.URL, req.Provider, req.AuthConfig)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": result.ID, "name": result.Name, "url": result.URL,
		"provider": result.Provider, "created_at": result.CreatedAt,
	})
}

// List 列出镜像仓库
func (h *RegistryHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.svc.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询失败")
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 获取单个镜像仓库
func (h *RegistryHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	result, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "镜像仓库不存在")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update 更新镜像仓库
func (h *RegistryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Name       string  `json:"name"`
		URL        string  `json:"url"`
		AuthConfig *string `json:"auth_config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	result, err := h.svc.Update(uint(id), req.Name, req.URL, req.AuthConfig)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete 删除镜像仓库
func (h *RegistryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterRegistryRoutes 注册镜像仓库路由
func RegisterRegistryRoutes(r *gin.RouterGroup, h *RegistryHandler) {
	regs := r.Group("/registries")
	{
		regs.GET("", h.List)
		regs.POST("", h.Create)
		regs.GET("/:id", h.Get)
		regs.PUT("/:id", h.Update)
		regs.DELETE("/:id", h.Delete)
	}
}

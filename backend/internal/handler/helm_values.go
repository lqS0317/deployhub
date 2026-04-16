package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"

	"github.com/gin-gonic/gin"
)

// HelmValuesHandler Helm Values 处理器
type HelmValuesHandler struct {
	repo repository.HelmValuesRepository
}

func NewHelmValuesHandler(repo repository.HelmValuesRepository) *HelmValuesHandler {
	return &HelmValuesHandler{repo: repo}
}

// ListValues 获取服务在各集群的 values 列表
func (h *HelmValuesHandler) ListValues(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	vals, err := h.repo.ListByService(uint(serviceID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询 Helm Values 失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": vals})
}

type updateHelmValuesRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateValues 编辑某集群的 values
func (h *HelmValuesHandler) UpdateValues(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	clusterID, _ := strconv.ParseUint(c.Param("cluster_id"), 10, 32)

	var req updateHelmValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)
	hv := &model.HelmValues{
		ServiceID: uint(serviceID),
		ClusterID: uint(clusterID),
		Content:   req.Content,
		UpdatedBy: &userID,
	}

	if err := h.repo.Upsert(hv); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "保存 Helm Values 失败")
		return
	}

	result, _ := h.repo.FindByServiceAndCluster(uint(serviceID), uint(clusterID))
	c.JSON(http.StatusOK, result)
}

// RegisterHelmValuesRoutes 注册 Helm Values 路由
func RegisterHelmValuesRoutes(r *gin.RouterGroup, h *HelmValuesHandler) {
	r.GET("/services/:id/helm-values", h.ListValues)
	r.PUT("/services/:id/helm-values/:cluster_id", h.UpdateValues)
}

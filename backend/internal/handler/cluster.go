package handler

import (
	"net/http"
	"strconv"
	"strings"

	"deployhub/internal/pkg"
	"deployhub/internal/service/cluster"

	"github.com/gin-gonic/gin"
)

// ClusterHandler 集群处理器
type ClusterHandler struct {
	clusterSvc *cluster.ClusterService
	clientPool *cluster.ClientsetPool
}

// NewClusterHandler 创建集群处理器
func NewClusterHandler(clusterSvc *cluster.ClusterService, clientPool *cluster.ClientsetPool) *ClusterHandler {
	return &ClusterHandler{clusterSvc: clusterSvc, clientPool: clientPool}
}

type createClusterRequest struct {
	Name                string `json:"name" binding:"required"`
	DisplayName         string `json:"display_name"`
	Env                 string `json:"env" binding:"required"`
	APIServer           string `json:"api_server"`
	Kubeconfig          string `json:"kubeconfig" binding:"required"`
	HelmServiceAccount  string `json:"helm_service_account"`
	BuildServiceAccount string `json:"build_service_account"`
}

// Create 创建集群
func (h *ClusterHandler) Create(c *gin.Context) {
	var req createClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	result, err := h.clusterSvc.Create(req.Name, req.DisplayName, req.Env, req.APIServer, req.Kubeconfig, req.HelmServiceAccount, req.BuildServiceAccount)
	if err != nil {
		if strings.Contains(err.Error(), "已存在") {
			pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		} else {
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.ID, "name": result.Name, "display_name": result.DisplayName,
		"env": result.Env, "api_server": result.APIServer, "created_at": result.CreatedAt,
	})
}

// List 列出集群
func (h *ClusterHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.clusterSvc.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询集群列表失败")
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 获取集群详情
func (h *ClusterHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	result, err := h.clusterSvc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "集群不存在")
		return
	}

	c.JSON(http.StatusOK, result)
}

type updateClusterRequest struct {
	DisplayName         string  `json:"display_name"`
	Env                 string  `json:"env"`
	APIServer           string  `json:"api_server"`
	Kubeconfig          *string `json:"kubeconfig"`
	HelmServiceAccount  *string `json:"helm_service_account"`
	BuildServiceAccount *string `json:"build_service_account"`
}

// Update 更新集群
func (h *ClusterHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	var req updateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	result, err := h.clusterSvc.Update(uint(id), req.DisplayName, req.Env, req.APIServer, req.Kubeconfig, req.HelmServiceAccount, req.BuildServiceAccount)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	if req.Kubeconfig != nil {
		h.clientPool.InvalidateCache(uint(id))
	}

	c.JSON(http.StatusOK, result)
}

// Delete 删除集群
func (h *ClusterHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	if err := h.clusterSvc.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	h.clientPool.InvalidateCache(uint(id))
	c.Status(http.StatusNoContent)
}

// TestConnection 测试集群连接
func (h *ClusterHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	cs, err := h.clientPool.GetClientset(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	result := cluster.TestConnection(cs)
	c.JSON(http.StatusOK, result)
}

// RegisterClusterRoutes 注册集群路由
func RegisterClusterRoutes(r *gin.RouterGroup, h *ClusterHandler) {
	clusters := r.Group("/clusters")
	{
		clusters.GET("", h.List)
		clusters.POST("", h.Create)
		clusters.GET("/:id", h.Get)
		clusters.PUT("/:id", h.Update)
		clusters.DELETE("/:id", h.Delete)
		clusters.POST("/:id/test", h.TestConnection)
	}
}

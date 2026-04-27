package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"
	"deployhub/internal/service/cluster"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceHandler 集群命名空间管理处理器
type NamespaceHandler struct {
	nsRepo     repository.ClusterNamespaceRepository
	clientPool *cluster.ClientsetPool
}

func NewNamespaceHandler(nsRepo repository.ClusterNamespaceRepository, clientPool *cluster.ClientsetPool) *NamespaceHandler {
	return &NamespaceHandler{nsRepo: nsRepo, clientPool: clientPool}
}

func (h *NamespaceHandler) List(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	nss, err := h.nsRepo.ListByCluster(uint(clusterID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询 namespace 列表失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": nss})
}

type createNamespaceRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

func (h *NamespaceHandler) Create(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req createNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	ns := &model.ClusterNamespace{
		ClusterID: uint(clusterID),
		Namespace: req.Namespace,
		IsDefault: req.IsDefault,
	}
	if err := h.nsRepo.Create(ns); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, "namespace 已存在")
		return
	}
	c.JSON(http.StatusCreated, ns)
}

func (h *NamespaceHandler) Delete(c *gin.Context) {
	nsID, _ := strconv.ParseUint(c.Param("ns_id"), 10, 32)
	if err := h.nsRepo.Delete(uint(nsID)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// Sync 通过 client-go 动态加载集群的所有 namespace
func (h *NamespaceHandler) Sync(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	clientset, err := h.clientPool.GetClientset(uint(clusterID))
	if err != nil {
		pkg.Error(c, http.StatusBadGateway, "CLUSTER_ERROR", "连接集群失败")
		return
	}

	nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		pkg.Error(c, http.StatusBadGateway, "CLUSTER_ERROR", "获取 namespace 列表失败")
		return
	}

	var added []string
	for _, ns := range nsList.Items {
		existing, err := h.nsRepo.FindByClusterAndNamespace(uint(clusterID), ns.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询 namespace 映射失败")
			return
		}
		if existing != nil {
			continue
		}
		if err := h.nsRepo.Create(&model.ClusterNamespace{
			ClusterID: uint(clusterID),
			Namespace: ns.Name,
		}); err != nil {
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "创建 namespace 映射失败")
			return
		}
		added = append(added, ns.Name)
	}

	c.JSON(http.StatusOK, gin.H{"synced": len(added), "namespaces": added})
}

// RegisterNamespaceRoutes 注册 namespace 路由
func RegisterNamespaceRoutes(r *gin.RouterGroup, h *NamespaceHandler) {
	r.GET("/clusters/:id/namespaces", h.List)
	r.POST("/clusters/:id/namespaces", middleware.AdminOnly(), h.Create)
	r.DELETE("/clusters/:id/namespaces/:ns_id", middleware.AdminOnly(), h.Delete)
	r.POST("/clusters/:id/namespaces/sync", middleware.AdminOnly(), h.Sync)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/service/svc"

	"gorm.io/datatypes"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	svcService       *svc.ServiceService
	memberService    *svc.MemberService
	effectiveRoleSvc *svc.EffectiveRoleService
}

func NewServiceHandler(svcService *svc.ServiceService, memberService *svc.MemberService, effectiveRoleSvc *svc.EffectiveRoleService) *ServiceHandler {
	return &ServiceHandler{svcService: svcService, memberService: memberService, effectiveRoleSvc: effectiveRoleSvc}
}

type createServiceRequest struct {
	Name            string `json:"name" binding:"required"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ServiceType     string `json:"service_type"`
	Language        string `json:"language"`
	LanguageVersion string `json:"language_version"`
	GitRepoID       uint   `json:"git_repo_id" binding:"required"`
	GitBranch       string `json:"git_branch"`
	// 以下字段为兼容旧数据保留，创建时非必填
	DockerfilePath  string `json:"dockerfile_path"`
	RegistryID      *uint  `json:"registry_id"`
	ImageRepo       string `json:"image_repo"`
	ClusterID       *uint  `json:"cluster_id"`
	Namespace       string `json:"namespace"`
	Replicas        int    `json:"replicas"`
	Port            int    `json:"port"`
	HealthCheckPath string `json:"health_check_path"`
	CPURequest      string `json:"cpu_request"`
	MemRequest      string `json:"mem_request"`
	CPULimit        string `json:"cpu_limit"`
	MemLimit        string `json:"mem_limit"`
	DeployType      string `json:"deploy_type"`
	HelmRepoID      *uint  `json:"helm_repo_id"`
	HelmChartPath   string `json:"helm_chart_path"`
	HelmValuesPath  string `json:"helm_values_path"`
	HelmReleaseName string `json:"helm_release_name"`
	HelmChartBranch string `json:"helm_chart_branch"`
	WorkloadType    string `json:"workload_type"`
	// 运行时默认配置（direct 部署类型用）
	DefaultPort                int             `json:"default_port"`
	DefaultReplicas            int             `json:"default_replicas"`
	DefaultCPURequest          string          `json:"default_cpu_request"`
	DefaultMemRequest          string          `json:"default_mem_request"`
	DefaultCPULimit            string          `json:"default_cpu_limit"`
	DefaultMemLimit            string          `json:"default_mem_limit"`
	DefaultCommand             json.RawMessage `json:"default_command"`
	DefaultArgs                json.RawMessage `json:"default_args"`
	DefaultWorkloadType        string          `json:"default_workload_type"`
	DefaultLivenessProbe       json.RawMessage `json:"default_liveness_probe"`
	DefaultReadinessProbe      json.RawMessage `json:"default_readiness_probe"`
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	userID := middleware.GetUserID(c)
	deployType := req.DeployType
	if deployType == "" {
		deployType = "direct"
	}
	svcModel := &model.Service{
		Name: req.Name, DisplayName: req.DisplayName, Description: req.Description,
		GitRepoID: req.GitRepoID, GitBranch: req.GitBranch, DockerfilePath: req.DockerfilePath,
		RegistryID: req.RegistryID, ImageRepo: req.ImageRepo, ClusterID: req.ClusterID,
		Namespace: req.Namespace, Replicas: req.Replicas, Port: req.Port,
		HealthCheckPath: req.HealthCheckPath, CPURequest: req.CPURequest, MemRequest: req.MemRequest,
		CPULimit: req.CPULimit, MemLimit: req.MemLimit, OwnerID: userID,
		ServiceType: req.ServiceType, Language: req.Language, LanguageVersion: req.LanguageVersion,
		DeployType: deployType, HelmRepoID: req.HelmRepoID,
		HelmChartPath: req.HelmChartPath, HelmValuesPath: req.HelmValuesPath,
		HelmReleaseName: req.HelmReleaseName, HelmChartBranch: req.HelmChartBranch,
		DefaultPort: req.DefaultPort, DefaultReplicas: maxInt(req.DefaultReplicas, 1),
		DefaultCPURequest: req.DefaultCPURequest, DefaultMemRequest: req.DefaultMemRequest,
		DefaultCPULimit: req.DefaultCPULimit, DefaultMemLimit: req.DefaultMemLimit,
		DefaultCommand: toDataJSON(req.DefaultCommand), DefaultArgs: toDataJSON(req.DefaultArgs),
		DefaultWorkloadType: req.DefaultWorkloadType,
		DefaultLivenessProbe: toDataJSON(req.DefaultLivenessProbe), DefaultReadinessProbe: toDataJSON(req.DefaultReadinessProbe),
	}
	if req.WorkloadType == "statefulset" {
		svcModel.WorkloadType = "statefulset"
	} else {
		svcModel.WorkloadType = "deployment"
	}
	if svcModel.GitBranch == "" {
		svcModel.GitBranch = "main"
	}
	if svcModel.DockerfilePath == "" {
		svcModel.DockerfilePath = "./Dockerfile"
	}
	if svcModel.Namespace == "" {
		svcModel.Namespace = "default"
	}
	if svcModel.Replicas == 0 {
		svcModel.Replicas = 1
	}

	result, err := h.svcService.Create(svcModel)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *ServiceHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.svcService.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询失败")
		return
	}

	// 非 Admin 用户按有效权限过滤，只返回有权限的 Service
	userRole := middleware.GetUserRole(c)
	if userRole != "admin" && h.effectiveRoleSvc != nil {
		userID := middleware.GetUserID(c)
		var filtered []model.Service
		for _, s := range items {
			role, _ := h.effectiveRoleSvc.GetEffectiveRole(userID, s.ID)
			if role != "" {
				filtered = append(filtered, s)
			}
		}
		pkg.Paginated(c, filtered, int64(len(filtered)), page, pageSize)
		return
	}

	pkg.Paginated(c, items, total, page, pageSize)
}

func (h *ServiceHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	result, err := h.svcService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "服务不存在")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	existing, err := h.svcService.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "服务不存在")
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	result, err := h.svcService.Update(existing)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svcService.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ServiceHandler) ListMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	members, err := h.memberService.ListMembers(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询成员失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *ServiceHandler) AddMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	if err := h.memberService.AddMember(uint(id), req.UserID, req.Role); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": req.UserID, "role": req.Role})
}

func (h *ServiceHandler) UpdateMemberRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	memberID, _ := strconv.ParseUint(c.Param("member_id"), 10, 32)
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	if err := h.memberService.UpdateRole(uint(id), uint(memberID), req.Role); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": req.Role})
}

func (h *ServiceHandler) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	memberID, _ := strconv.ParseUint(c.Param("member_id"), 10, 32)
	if err := h.memberService.RemoveMember(uint(id), uint(memberID)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ImportPreview 导入预览
func (h *ServiceHandler) ImportPreview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "请上传文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "打开文件失败")
		return
	}
	defer f.Close()
	results, err := svc.ParseImportFile(f, file.Header.Get("Content-Type"))
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"preview": results})
}

// ImportConfirm 确认导入
func (h *ServiceHandler) ImportConfirm(c *gin.Context) {
	var req struct {
		Services []svc.ImportedService `json:"services" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	userID := middleware.GetUserID(c)
	var models []*model.Service
	for _, s := range req.Services {
		models = append(models, &model.Service{
			Name: s.Name, ImageRepo: s.Image, Replicas: s.Replicas, Port: s.Port,
			Namespace: s.Namespace, OwnerID: userID, GitBranch: "main",
			DockerfilePath: "./Dockerfile",
		})
	}
	if err := h.svcService.BatchCreate(models); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": len(models)})
}

// RegisterServiceRoutes 注册服务路由
func RegisterServiceRoutes(r *gin.RouterGroup, h *ServiceHandler, rbacSvc *svc.RBACService) {
	services := r.Group("/services")
	{
		services.GET("", h.List)
		services.POST("", h.Create)
		services.POST("/import/preview", h.ImportPreview)
		services.POST("/import", h.ImportConfirm)
		services.GET("/:id", middleware.ServiceRBAC(rbacSvc, "viewer"), h.Get)
		services.PUT("/:id", middleware.ServiceRBAC(rbacSvc, "developer"), h.Update)
		services.DELETE("/:id", middleware.ServiceRBAC(rbacSvc, "owner"), h.Delete)
		services.GET("/:id/members", middleware.ServiceRBAC(rbacSvc, "viewer"), h.ListMembers)
		services.POST("/:id/members", middleware.ServiceRBAC(rbacSvc, "owner"), h.AddMember)
		services.PUT("/:id/members/:member_id", middleware.ServiceRBAC(rbacSvc, "owner"), h.UpdateMemberRole)
		services.DELETE("/:id/members/:member_id", middleware.ServiceRBAC(rbacSvc, "owner"), h.RemoveMember)
	}
}

func toDataJSON(raw json.RawMessage) datatypes.JSON {
	if len(raw) == 0 {
		return nil
	}
	return datatypes.JSON(raw)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

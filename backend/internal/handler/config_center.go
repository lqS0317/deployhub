package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/configcenter"

	"github.com/gin-gonic/gin"
)

// ConfigCenterHandler 配置中心处理器
type ConfigCenterHandler struct {
	configSvc *configcenter.ConfigService
	permSvc   *configcenter.ConfigPermissionService
}

// NewConfigCenterHandler 创建配置中心处理器
func NewConfigCenterHandler(
	configSvc *configcenter.ConfigService,
	permSvc *configcenter.ConfigPermissionService,
) *ConfigCenterHandler {
	return &ConfigCenterHandler{configSvc: configSvc, permSvc: permSvc}
}

// ---- 请求结构体 ----

type ccCreateEntryRequest struct {
	ClusterID  uint   `json:"cluster_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	ConfigType string `json:"config_type" binding:"required"`
	Format     string `json:"format" binding:"required"`
	MountPath  string `json:"mount_path"`
}

type ccUpdateEntryRequest struct {
	Name         string `json:"name"`
	MountPath    string `json:"mount_path"`
	DraftContent string `json:"draft_content"`
}

type ccCreateItemRequest struct {
	Key     string `json:"key" binding:"required"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type ccUpdateItemRequest struct {
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type ccSaveDraftRequest struct {
	Content string `json:"content" binding:"required"`
}

type ccPublishRequest struct {
	Comment string `json:"comment"`
}

type ccRollbackRequest struct {
	Version int    `json:"version" binding:"required"`
	Comment string `json:"comment"`
}

type ccGrantPermissionRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// parseServiceID 解析服务 ID
func parseServiceID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的服务 ID")
		return 0, false
	}
	return uint(id), true
}

// parseEntryID 解析配置条目 ID
func parseEntryID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("entry_id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置条目 ID")
		return 0, false
	}
	return uint(id), true
}

// getEntryPermissionContext 从 entry 获取权限校验所需的 serviceID 和 clusterID
func (h *ConfigCenterHandler) getEntryPermissionContext(c *gin.Context, entryID uint) (serviceID, clusterID uint, ok bool) {
	entry, err := h.configSvc.GetEntry(entryID)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, err.Error())
		return 0, 0, false
	}
	return entry.ServiceID, entry.ClusterID, true
}

// ---- Entry 操作 ----

// ListPublishedEntries 列出该服务+集群下所有已发布的配置条目（供部署预览用）
func (h *ConfigCenterHandler) ListPublishedEntries(c *gin.Context) {
	serviceID, ok := parseServiceID(c)
	if !ok {
		return
	}
	clusterIDStr := c.Query("cluster_id")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 cluster_id")
		return
	}

	entries, err := h.configSvc.ListEntries(serviceID, uint(clusterID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	// 过滤出有已发布版本的条目
	type publishedEntry struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		ConfigType string `json:"config_type"`
		Format     string `json:"format"`
		MountPath  string `json:"mount_path"`
		Version    int    `json:"version"`
	}
	var result []publishedEntry
	for _, e := range entries {
		release, _ := h.configSvc.GetPublishedSnapshot(e.ID)
		if release != nil {
			result = append(result, publishedEntry{
				ID: e.ID, Name: e.Name, ConfigType: e.ConfigType,
				Format: e.Format, MountPath: e.MountPath, Version: release.Version,
			})
		}
	}
	pkg.Success(c, http.StatusOK, result)
}

// ListEntries 列出配置条目
func (h *ConfigCenterHandler) ListEntries(c *gin.Context) {
	serviceID, ok := parseServiceID(c)
	if !ok {
		return
	}
	clusterIDStr := c.Query("cluster_id")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 cluster_id")
		return
	}

	entries, err := h.configSvc.ListEntries(serviceID, uint(clusterID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, entries)
}

// CreateEntry 创建配置条目
func (h *ConfigCenterHandler) CreateEntry(c *gin.Context) {
	serviceID, ok := parseServiceID(c)
	if !ok {
		return
	}

	var req ccCreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, req.ClusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	entry, err := h.configSvc.CreateEntry(serviceID, req.ClusterID, req.Name, req.ConfigType, req.Format, req.MountPath)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}
	pkg.Success(c, http.StatusCreated, entry)
}

// GetEntry 获取配置条目详情
func (h *ConfigCenterHandler) GetEntry(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	entry, err := h.configSvc.GetEntry(entryID)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, entry)
}

// UpdateEntry 更新配置条目
func (h *ConfigCenterHandler) UpdateEntry(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	var req ccUpdateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	entry, err := h.configSvc.UpdateEntry(entryID, req.Name, req.DraftContent)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, entry)
}

// DeleteEntry 删除配置条目
func (h *ConfigCenterHandler) DeleteEntry(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "publisher"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	if err := h.configSvc.DeleteEntry(entryID); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"message": "删除成功"})
}

// ---- Item 操作 ----

// ListItems 列出配置项
func (h *ConfigCenterHandler) ListItems(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "viewer"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	decrypt := false
	if permErr := h.permSvc.CheckPermission(serviceID, clusterID, userID, "publisher"); permErr == nil {
		decrypt = true
	}

	items, err := h.configSvc.ListItems(entryID, decrypt)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, items)
}

// CreateItem 创建配置项
func (h *ConfigCenterHandler) CreateItem(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	var req ccCreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	item, err := h.configSvc.CreateItem(entryID, req.Key, req.Value, req.Comment)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}
	pkg.Success(c, http.StatusCreated, item)
}

// UpdateItem 更新配置项
func (h *ConfigCenterHandler) UpdateItem(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置项 ID")
		return
	}

	var req ccUpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	item, err := h.configSvc.UpdateItem(uint(itemID), req.Value, req.Comment)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, item)
}

// DeleteItem 删除配置项
func (h *ConfigCenterHandler) DeleteItem(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的配置项 ID")
		return
	}

	if err := h.configSvc.DeleteItem(uint(itemID)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"message": "删除成功"})
}

// ---- Draft 操作 ----

// GetDraft 获取草稿
func (h *ConfigCenterHandler) GetDraft(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "viewer"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	content, err := h.configSvc.GetDraft(entryID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"content": content})
}

// SaveDraft 保存草稿
func (h *ConfigCenterHandler) SaveDraft(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "editor"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	var req ccSaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.configSvc.SaveDraft(entryID, req.Content); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"message": "保存成功"})
}

// ---- Publish / Rollback ----

// Publish 发布配置
func (h *ConfigCenterHandler) Publish(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "publisher"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	var req ccPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	release, err := h.configSvc.Publish(entryID, userID, req.Comment)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusCreated, release)
}

// Rollback 回滚配置
func (h *ConfigCenterHandler) Rollback(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	serviceID, clusterID, ok := h.getEntryPermissionContext(c, entryID)
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.permSvc.CheckPermission(serviceID, clusterID, userID, "publisher"); err != nil {
		pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, err.Error())
		return
	}

	var req ccRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	release, err := h.configSvc.Rollback(entryID, userID, req.Version, req.Comment)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusCreated, release)
}

// ---- Release 列表 ----

// ListReleases 列出发布历史
func (h *ConfigCenterHandler) ListReleases(c *gin.Context) {
	entryID, ok := parseEntryID(c)
	if !ok {
		return
	}

	releases, err := h.configSvc.ListReleases(entryID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, releases)
}

// ---- Permission 操作 ----

// ListPermissions 列出服务配置权限
func (h *ConfigCenterHandler) ListPermissions(c *gin.Context) {
	serviceID, ok := parseServiceID(c)
	if !ok {
		return
	}

	perms, err := h.permSvc.ListPermissions(serviceID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, perms)
}

// GrantPermission 授予权限
func (h *ConfigCenterHandler) GrantPermission(c *gin.Context) {
	serviceID, ok := parseServiceID(c)
	if !ok {
		return
	}
	cidStr := c.Param("cid")
	clusterID, err := strconv.ParseUint(cidStr, 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的集群 ID")
		return
	}

	var req ccGrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.permSvc.GrantPermission(serviceID, uint(clusterID), req.UserID, req.Role); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"message": "授权成功"})
}

// RevokePermission 撤销权限
func (h *ConfigCenterHandler) RevokePermission(c *gin.Context) {
	permID, err := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的权限 ID")
		return
	}

	if err := h.permSvc.RevokePermission(uint(permID)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	pkg.Success(c, http.StatusOK, gin.H{"message": "撤销成功"})
}

// RegisterConfigCenterRoutes 注册配置中心路由
func RegisterConfigCenterRoutes(r *gin.RouterGroup, h *ConfigCenterHandler) {
	// 配置条目（按服务维度）
	svc := r.Group("/services/:id")
	{
		svc.GET("/config-entries", h.ListEntries)
		svc.POST("/config-entries", h.CreateEntry)
	}

	// 部署预览：列出该服务+集群下所有已发布的配置条目
	svc.GET("/config-entries/published", h.ListPublishedEntries)

	// 配置条目操作（按条目维度）
	entries := r.Group("/config-entries/:entry_id")
	{
		entries.GET("", h.GetEntry)
		entries.PUT("", h.UpdateEntry)
		entries.DELETE("", h.DeleteEntry)
		entries.GET("/items", h.ListItems)
		entries.POST("/items", h.CreateItem)
		entries.PUT("/items/:item_id", h.UpdateItem)
		entries.DELETE("/items/:item_id", h.DeleteItem)
		entries.GET("/draft", h.GetDraft)
		entries.PUT("/draft", h.SaveDraft)
		entries.POST("/release", h.Publish)
		entries.POST("/rollback", h.Rollback)
		entries.GET("/releases", h.ListReleases)
	}

	// 权限路由（保持旧路径兼容）
	perms := r.Group("/services/:id/configs/:cid")
	{
		perms.GET("/permissions", h.ListPermissions)
		perms.POST("/permissions", h.GrantPermission)
		perms.DELETE("/permissions/:pid", h.RevokePermission)
	}
}

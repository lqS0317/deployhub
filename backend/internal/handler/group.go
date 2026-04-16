package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	groupSvc "deployhub/internal/service/group"

	"github.com/gin-gonic/gin"
)

// GroupHandler 组管理处理器
type GroupHandler struct {
	svc *groupSvc.GroupService
}

func NewGroupHandler(svc *groupSvc.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

type createGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	userID := middleware.GetUserID(c)
	g, err := h.svc.Create(req.Name, req.Description, userID)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (h *GroupHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.svc.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询组列表失败")
		return
	}
	type groupResp struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		CreatedBy       uint   `json:"created_by"`
		MemberCount     int    `json:"member_count"`
		PermissionCount int    `json:"permission_count"`
		CreatedAt       string `json:"created_at"`
	}
	var resp []groupResp
	for _, g := range items {
		resp = append(resp, groupResp{
			ID:              g.ID,
			Name:            g.Name,
			Description:     g.Description,
			CreatedBy:       g.CreatedBy,
			MemberCount:     h.svc.GetMemberCount(g.ID),
			PermissionCount: h.svc.GetPermissionCount(g.ID),
			CreatedAt:       g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	pkg.Paginated(c, resp, total, page, pageSize)
}

func (h *GroupHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	g, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "组不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":               g.ID,
		"name":             g.Name,
		"description":      g.Description,
		"created_by":       g.CreatedBy,
		"member_count":     h.svc.GetMemberCount(g.ID),
		"permission_count": h.svc.GetPermissionCount(g.ID),
		"created_at":       g.CreatedAt,
		"updated_at":       g.UpdatedAt,
	})
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	g, err := h.svc.Update(uint(id), req.Name, req.Description)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 成员管理 ---

type addMembersRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}

func (h *GroupHandler) ListMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	members, err := h.svc.ListMembers(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询成员列表失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *GroupHandler) AddMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req addMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	added, err := h.svc.AddMembers(uint(id), req.UserIDs)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"items": added})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID, _ := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err := h.svc.RemoveMember(uint(id), uint(userID)); err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "成员不存在")
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 权限管理 ---

type addPermissionRequest struct {
	ServiceID uint   `json:"service_id" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

func (h *GroupHandler) ListPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	perms, err := h.svc.ListPermissions(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询权限列表失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": perms})
}

func (h *GroupHandler) AddPermission(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req addPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	p, err := h.svc.AddPermission(uint(id), req.ServiceID, req.Role)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, p)
}

type updatePermissionRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *GroupHandler) UpdatePermission(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 32)
	var req updatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	if err := h.svc.UpdatePermission(uint(pid), req.Role); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "权限已更新"})
}

func (h *GroupHandler) RemovePermission(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 32)
	if err := h.svc.RemovePermission(uint(pid)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterGroupRoutes 注册组管理路由（Admin only）
func RegisterGroupRoutes(r *gin.RouterGroup, h *GroupHandler) {
	groups := r.Group("/groups", middleware.AdminOnly())
	{
		groups.GET("", h.List)
		groups.POST("", h.Create)
		groups.GET("/:id", h.Get)
		groups.PUT("/:id", h.Update)
		groups.DELETE("/:id", h.Delete)
		groups.GET("/:id/members", h.ListMembers)
		groups.POST("/:id/members", h.AddMembers)
		groups.DELETE("/:id/members/:user_id", h.RemoveMember)
		groups.GET("/:id/permissions", h.ListPermissions)
		groups.POST("/:id/permissions", h.AddPermission)
		groups.PUT("/:id/permissions/:pid", h.UpdatePermission)
		groups.DELETE("/:id/permissions/:pid", h.RemovePermission)
	}
}

package handler

import (
	"net/http"
	"strconv"
	"strings"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"
	"deployhub/internal/service/auth"

	"github.com/gin-gonic/gin"
)

// UserAdminHandler 管理员用户管理处理器
type UserAdminHandler struct {
	userRepo        repository.UserRepository
	authSvc         *auth.AuthService
	groupMemberRepo repository.GroupMemberRepository
}

// NewUserAdminHandler 创建管理员用户管理处理器
func NewUserAdminHandler(userRepo repository.UserRepository, authSvc *auth.AuthService, groupMemberRepo repository.GroupMemberRepository) *UserAdminHandler {
	return &UserAdminHandler{userRepo: userRepo, authSvc: authSvc, groupMemberRepo: groupMemberRepo}
}

// List 列出所有用户（分页）
func (h *UserAdminHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)

	users, total, err := h.userRepo.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询用户列表失败")
		return
	}

	pkg.Paginated(c, users, total, page, pageSize)
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// UpdateRole 更新用户角色
func (h *UserAdminHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的用户 ID")
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败: role 须为 admin 或 member")
		return
	}

	// 防止管理员修改自身角色导致无管理员
	currentUserID := middleware.GetUserID(c)
	if uint(id) == currentUserID {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "不能修改自己的角色")
		return
	}

	if _, err := h.userRepo.FindByID(uint(id)); err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "用户不存在")
		return
	}

	if err := h.userRepo.UpdateRole(uint(id), req.Role); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "更新角色失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色已更新"})
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UpdateStatus 更新用户状态
func (h *UserAdminHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的用户 ID")
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败: status 须为 active 或 disabled")
		return
	}

	// 防止管理员禁用自己
	currentUserID := middleware.GetUserID(c)
	if uint(id) == currentUserID && req.Status == "disabled" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "不能禁用自己的账户")
		return
	}

	if _, err := h.userRepo.FindByID(uint(id)); err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "用户不存在")
		return
	}

	if err := h.userRepo.UpdateStatus(uint(id), req.Status); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "更新状态失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin member"`
}

// CreateUser Admin 创建用户
func (h *UserAdminHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	user, err := h.authSvc.Register(req.Username, req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "已存在") {
			pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		} else {
			pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		}
		return
	}

	// 注册后如果角色不是 member（默认），则更新角色
	if req.Role != "member" {
		_ = h.userRepo.UpdateRole(user.ID, req.Role)
		user.Role = req.Role
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": user.ID, "username": user.Username, "email": user.Email, "role": user.Role,
	})
}

// GetUserGroups 获取用户所属的组列表
func (h *UserAdminHandler) GetUserGroups(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	members, err := h.groupMemberRepo.ListByUser(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询用户组列表失败")
		return
	}

	type groupInfo struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	var groups []groupInfo
	for _, m := range members {
		if m.Group != nil {
			groups = append(groups, groupInfo{ID: m.Group.ID, Name: m.Group.Name})
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": groups})
}

// RegisterUserAdminRoutes 注册管理员用户管理路由（仅管理员可访问）
func RegisterUserAdminRoutes(r *gin.RouterGroup, h *UserAdminHandler) {
	users := r.Group("/users", middleware.AdminOnly())
	{
		users.GET("", h.List)
		users.POST("", h.CreateUser)
		users.PUT("/:id/role", h.UpdateRole)
		users.PUT("/:id/status", h.UpdateStatus)
		users.GET("/:id/groups", h.GetUserGroups)
	}
}

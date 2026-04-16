package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/auth"
	"deployhub/internal/service/storage"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authSvc    *auth.AuthService
	jwtSvc     *auth.JWTService
	storageSvc storage.StorageService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc *auth.AuthService, jwtSvc *auth.JWTService, storageSvc storage.StorageService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, jwtSvc: jwtSvc, storageSvc: storageSvc}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	user, err := h.authSvc.Register(req.Username, req.Email, req.Password)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	token, user, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		pkg.Error(c, http.StatusUnauthorized, pkg.CodeUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// GetMe 获取当前用户完整信息（含脱敏手机号）
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	profile, err := h.authSvc.GetUserProfile(userID)
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "用户不存在")
		return
	}

	c.JSON(http.StatusOK, profile)
}

type updateProfileRequest struct {
	Nickname *string `json:"nickname"`
	Phone    *string `json:"phone"`
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)
	err := h.authSvc.UpdateProfile(userID, auth.UpdateProfileInput{
		Nickname: req.Nickname,
		Phone:    req.Phone,
	})
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "资料更新成功"})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "旧密码错误") {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, err.Error())
			return
		}
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

var allowedImageTypes = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

const maxAvatarSize = 2 * 1024 * 1024 // 2MB

// UploadAvatar 上传头像
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "请选择头像文件")
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "头像文件不能超过 2MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageTypes[ext] {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "仅支持 jpg/png/gif/webp 格式")
		return
	}

	contentType := fmt.Sprintf("image/%s", strings.TrimPrefix(ext, "."))
	if ext == ".jpg" {
		contentType = "image/jpeg"
	}

	objectKey, err := h.storageSvc.Upload(c.Request.Context(), "avatars", header.Filename, file, contentType)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "头像上传失败")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.authSvc.UpdateAvatar(userID, objectKey); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "更新头像失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar": objectKey})
}

// Logout 登出（前端清除 JWT，后端无状态记录）
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

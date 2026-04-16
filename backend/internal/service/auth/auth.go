package auth

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"

	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSvc    *JWTService
	cryptoSvc *crypto.CryptoService
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, jwtSvc *JWTService, cryptoSvc *crypto.CryptoService) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSvc:    jwtSvc,
		cryptoSvc: cryptoSvc,
	}
}

// Register 本地用户注册
func (s *AuthService) Register(username, email, password string) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, fmt.Errorf("用户名 %s 已存在", username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, fmt.Errorf("邮箱 %s 已存在", email)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询邮箱失败: %w", err)
	}

	// 哈希密码
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: &hash,
		Role:         "member",
		Status:       "active",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// Login 本地登录，返回 JWT 和用户信息
func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, fmt.Errorf("用户名或密码错误")
		}
		return "", nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查账号状态
	if user.Status == "disabled" {
		return "", nil, fmt.Errorf("账号已禁用")
	}

	// 校验密码
	if user.PasswordHash == nil {
		return "", nil, fmt.Errorf("该账号不支持密码登录")
	}
	if err := VerifyPassword(*user.PasswordHash, password); err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 签发 JWT
	token, err := s.jwtSvc.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, fmt.Errorf("签发令牌失败: %w", err)
	}

	return token, user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// UserProfileResponse 用户资料响应（含脱敏手机号）
type UserProfileResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Avatar    string `json:"avatar"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetUserProfile 获取完整用户资料（手机号脱敏）
func (s *AuthService) GetUserProfile(id uint) (*UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	phone := ""
	if user.PhoneEncrypted != "" {
		decrypted, err := s.cryptoSvc.Decrypt(user.PhoneEncrypted)
		if err == nil {
			phone = pkg.MaskPhone(decrypted)
		}
	}

	return &UserProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Avatar:    user.Avatar,
		Nickname:  user.Nickname,
		Phone:     phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// UpdateProfileInput 更新资料请求
type UpdateProfileInput struct {
	Nickname *string `json:"nickname"`
	Phone    *string `json:"phone"`
}

// UpdateProfile 更新用户资料（昵称和手机号）
func (s *AuthService) UpdateProfile(userID uint, input UpdateProfileInput) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}

	if input.Phone != nil {
		if *input.Phone == "" {
			user.PhoneEncrypted = ""
		} else {
			encrypted, err := s.cryptoSvc.Encrypt(*input.Phone)
			if err != nil {
				return fmt.Errorf("手机号加密失败: %w", err)
			}
			user.PhoneEncrypted = encrypted
		}
	}

	return s.userRepo.Update(user)
}

// ChangePassword 修改密码（需验证旧密码）
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	if user.PasswordHash == nil {
		return fmt.Errorf("该账号不支持密码登录")
	}

	if err := VerifyPassword(*user.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	user.PasswordHash = &hash
	return s.userRepo.Update(user)
}

// UpdateAvatar 更新用户头像 URL
func (s *AuthService) UpdateAvatar(userID uint, avatarURL string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	user.Avatar = avatarURL
	return s.userRepo.Update(user)
}

package auth

import (
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/service/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 测试用 AES 密钥（32 字节 = 64 位十六进制）
const testAESKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func newTestCryptoSvc(t *testing.T) *crypto.CryptoService {
	t.Helper()
	svc, err := crypto.NewCryptoService(testAESKey)
	require.NoError(t, err)
	return svc
}

// mockUserRepo 模拟用户仓库
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByOAuth(provider, oauthID string) (*model.User, error) {
	args := m.Called(provider, oauthID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepo) UpdateRole(id uint, role string) error {
	args := m.Called(id, role)
	return args.Error(0)
}

func (m *mockUserRepo) UpdateStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *mockUserRepo) FindByRole(role string) ([]model.User, error) { return nil, nil }

func TestRegister(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	repo.On("FindByUsername", "alice").Return(nil, gorm.ErrRecordNotFound)
	repo.On("FindByEmail", "alice@example.com").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Register("alice", "alice@example.com", "Password123!")
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "member", user.Role)
	repo.AssertExpectations(t)
}

func TestRegisterDuplicateUsername(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	existing := &model.User{Username: "alice"}
	repo.On("FindByUsername", "alice").Return(existing, nil)

	_, err := svc.Register("alice", "new@example.com", "Password123!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已存在")
}

func TestLogin(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	hash, _ := HashPassword("Password123!")
	user := &model.User{
		ID:           1,
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: &hash,
		Role:         "member",
		Status:       "active",
	}
	repo.On("FindByUsername", "alice").Return(user, nil)

	token, returnedUser, err := svc.Login("alice", "Password123!")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "alice", returnedUser.Username)
}

func TestLoginWrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	hash, _ := HashPassword("correct-password")
	user := &model.User{
		ID:           1,
		Username:     "alice",
		PasswordHash: &hash,
		Status:       "active",
	}
	repo.On("FindByUsername", "alice").Return(user, nil)

	_, _, err := svc.Login("alice", "wrong-password")
	assert.Error(t, err)
}

func TestLoginDisabledUser(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	hash, _ := HashPassword("Password123!")
	user := &model.User{
		ID:           1,
		Username:     "alice",
		PasswordHash: &hash,
		Status:       "disabled",
	}
	repo.On("FindByUsername", "alice").Return(user, nil)

	_, _, err := svc.Login("alice", "Password123!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已禁用")
}

func TestUpdateProfile(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	cryptoSvc := newTestCryptoSvc(t)
	svc := NewAuthService(repo, jwtSvc, cryptoSvc)

	user := &model.User{ID: 1, Username: "alice", Nickname: ""}
	repo.On("FindByID", uint(1)).Return(user, nil)
	repo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	nickname := "Alice"
	phone := "13812345678"
	err := svc.UpdateProfile(1, UpdateProfileInput{Nickname: &nickname, Phone: &phone})
	require.NoError(t, err)
	assert.Equal(t, "Alice", user.Nickname)
	assert.NotEmpty(t, user.PhoneEncrypted)
}

func TestChangePassword(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	hash, _ := HashPassword("OldPass123!")
	user := &model.User{ID: 1, PasswordHash: &hash}
	repo.On("FindByID", uint(1)).Return(user, nil)
	repo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	err := svc.ChangePassword(1, "OldPass123!", "NewPass456!")
	require.NoError(t, err)
}

func TestChangePasswordWrongOld(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc, newTestCryptoSvc(t))

	hash, _ := HashPassword("OldPass123!")
	user := &model.User{ID: 1, PasswordHash: &hash}
	repo.On("FindByID", uint(1)).Return(user, nil)

	err := svc.ChangePassword(1, "WrongPass!", "NewPass456!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "旧密码错误")
}

func TestGetUserProfile_WithPhone(t *testing.T) {
	repo := new(mockUserRepo)
	jwtSvc := NewJWTService("test-secret")
	cryptoSvc := newTestCryptoSvc(t)
	svc := NewAuthService(repo, jwtSvc, cryptoSvc)

	encrypted, _ := cryptoSvc.Encrypt("13812345678")
	user := &model.User{
		ID: 1, Username: "alice", Email: "alice@example.com",
		Role: "member", Nickname: "Alice", PhoneEncrypted: encrypted, Status: "active",
	}
	repo.On("FindByID", uint(1)).Return(user, nil)

	profile, err := svc.GetUserProfile(1)
	require.NoError(t, err)
	assert.Equal(t, "138****5678", profile.Phone)
	assert.Equal(t, "Alice", profile.Nickname)
}

// 确保 mockUserRepo 实现了接口（编译期检查）
var _ interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByOAuth(provider, oauthID string) (*model.User, error)
	Update(user *model.User) error
	List(page, pageSize int) ([]model.User, int64, error)
	UpdateRole(id uint, role string) error
	UpdateStatus(id uint, status string) error
	FindByRole(role string) ([]model.User, error)
} = (*mockUserRepo)(nil)

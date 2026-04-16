package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/auth"
	"deployhub/internal/service/crypto"
	"deployhub/internal/service/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testAESKeyHandler = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type mockUserRepoForHandler struct {
	mock.Mock
}

func (m *mockUserRepoForHandler) Create(user *model.User) error {
	return m.Called(user).Error(0)
}

func (m *mockUserRepoForHandler) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForHandler) FindByUsername(u string) (*model.User, error) {
	args := m.Called(u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForHandler) FindByEmail(e string) (*model.User, error) {
	args := m.Called(e)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForHandler) FindByOAuth(p, o string) (*model.User, error) {
	args := m.Called(p, o)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForHandler) Update(user *model.User) error {
	return m.Called(user).Error(0)
}

func (m *mockUserRepoForHandler) List(p, ps int) ([]model.User, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepoForHandler) UpdateRole(id uint, role string) error {
	return m.Called(id, role).Error(0)
}

func (m *mockUserRepoForHandler) UpdateStatus(id uint, status string) error {
	return m.Called(id, status).Error(0)
}

func (m *mockUserRepoForHandler) FindByRole(role string) ([]model.User, error) { return nil, nil }

var _ repository.UserRepository = (*mockUserRepoForHandler)(nil)

func setupAuthRouter() (*gin.Engine, *mockUserRepoForHandler, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	repo := new(mockUserRepoForHandler)
	jwtSvc := auth.NewJWTService("test-secret")
	cryptoSvc, _ := crypto.NewCryptoService(testAESKeyHandler)
	authSvc := auth.NewAuthService(repo, jwtSvc, cryptoSvc)
	storageSvc := storage.NewStorageService("", "", "", "", "")
	h := NewAuthHandler(authSvc, jwtSvc, storageSvc)

	r := gin.New()
	authGroup := r.Group("/api/v1/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.GET("/me", middleware.JWTAuth(jwtSvc), h.GetMe)
	authGroup.PUT("/profile", middleware.JWTAuth(jwtSvc), h.UpdateProfile)
	authGroup.PUT("/password", middleware.JWTAuth(jwtSvc), h.ChangePassword)
	authGroup.POST("/logout", middleware.JWTAuth(jwtSvc), h.Logout)
	return r, repo, jwtSvc
}

func TestRegisterHandler(t *testing.T) {
	r, repo, _ := setupAuthRouter()

	repo.On("FindByUsername", "alice").Return(nil, gorm.ErrRecordNotFound)
	repo.On("FindByEmail", "alice@example.com").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	body, _ := json.Marshal(map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Password123!",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLoginHandler(t *testing.T) {
	r, repo, _ := setupAuthRouter()

	hash, _ := auth.HashPassword("Password123!")
	user := &model.User{ID: 1, Username: "alice", Email: "alice@example.com", PasswordHash: &hash, Role: "member", Status: "active"}
	repo.On("FindByUsername", "alice").Return(user, nil)

	body, _ := json.Marshal(map[string]string{
		"username": "alice",
		"password": "Password123!",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["access_token"])
}

func TestGetMeHandler(t *testing.T) {
	r, repo, jwtSvc := setupAuthRouter()

	user := &model.User{ID: 1, Username: "alice", Email: "alice@example.com", Role: "member", Status: "active"}
	repo.On("FindByID", uint(1)).Return(user, nil)

	token, _ := jwtSvc.GenerateToken(1, "alice", "member")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

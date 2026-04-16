package registry

import (
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/service/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockRegistryRepo struct {
	mock.Mock
}

func (m *mockRegistryRepo) Create(r *model.Registry) error { return m.Called(r).Error(0) }
func (m *mockRegistryRepo) FindByID(id uint) (*model.Registry, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Registry), args.Error(1)
}
func (m *mockRegistryRepo) FindByName(name string) (*model.Registry, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Registry), args.Error(1)
}
func (m *mockRegistryRepo) Update(r *model.Registry) error { return m.Called(r).Error(0) }
func (m *mockRegistryRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockRegistryRepo) List(p, ps int) ([]model.Registry, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.Registry), args.Get(1).(int64), args.Error(2)
}

func TestCreateRegistry(t *testing.T) {
	repo := new(mockRegistryRepo)
	cryptoSvc, _ := crypto.NewCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	svc := NewRegistryService(repo, cryptoSvc)

	repo.On("FindByName", "my-harbor").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*model.Registry")).Return(nil)

	result, err := svc.Create("my-harbor", "https://harbor.example.com", "harbor", `{"username":"robot","password":"xxx"}`)
	require.NoError(t, err)
	assert.Equal(t, "my-harbor", result.Name)
	assert.NotEqual(t, `{"username":"robot","password":"xxx"}`, result.AuthConfigEncrypted)
}

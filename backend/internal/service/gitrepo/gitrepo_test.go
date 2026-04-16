package gitrepo

import (
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/service/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockGitRepoRepo struct {
	mock.Mock
}

func (m *mockGitRepoRepo) Create(r *model.GitRepo) error { return m.Called(r).Error(0) }
func (m *mockGitRepoRepo) FindByID(id uint) (*model.GitRepo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GitRepo), args.Error(1)
}
func (m *mockGitRepoRepo) FindByName(name string) (*model.GitRepo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GitRepo), args.Error(1)
}
func (m *mockGitRepoRepo) Update(r *model.GitRepo) error { return m.Called(r).Error(0) }
func (m *mockGitRepoRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockGitRepoRepo) List(p, ps int) ([]model.GitRepo, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.GitRepo), args.Get(1).(int64), args.Error(2)
}

func TestCreateGitRepo(t *testing.T) {
	repo := new(mockGitRepoRepo)
	cryptoSvc, _ := crypto.NewCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	svc := NewGitRepoService(repo, cryptoSvc)

	repo.On("FindByName", "my-repo").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*model.GitRepo")).Return(nil)

	result, err := svc.Create("my-repo", "https://github.com/org/repo.git", "github", "token", "ghp_xxxx")
	require.NoError(t, err)
	assert.Equal(t, "my-repo", result.Name)
	assert.NotEqual(t, "ghp_xxxx", result.CredentialEncrypted)
}

func TestCreateGitRepoDuplicate(t *testing.T) {
	repo := new(mockGitRepoRepo)
	cryptoSvc, _ := crypto.NewCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	svc := NewGitRepoService(repo, cryptoSvc)

	existing := &model.GitRepo{Name: "my-repo"}
	repo.On("FindByName", "my-repo").Return(existing, nil)

	_, err := svc.Create("my-repo", "url", "github", "token", "credential")
	assert.Error(t, err)
}

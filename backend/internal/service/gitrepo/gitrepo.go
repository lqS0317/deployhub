package gitrepo

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"

	"gorm.io/gorm"
)

// GitRepoService Git 仓库管理服务
type GitRepoService struct {
	repo      repository.GitRepoRepository
	cryptoSvc *crypto.CryptoService
}

// NewGitRepoService 创建 Git 仓库服务
func NewGitRepoService(repo repository.GitRepoRepository, cryptoSvc *crypto.CryptoService) *GitRepoService {
	return &GitRepoService{repo: repo, cryptoSvc: cryptoSvc}
}

// Create 创建 Git 仓库（凭证加密存储）
func (s *GitRepoService) Create(name, url, provider, authType, credential string) (*model.GitRepo, error) {
	if _, err := s.repo.FindByName(name); err == nil {
		return nil, fmt.Errorf("Git 仓库 %s 已存在", name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询 Git 仓库失败: %w", err)
	}

	encrypted, err := s.cryptoSvc.Encrypt(credential)
	if err != nil {
		return nil, fmt.Errorf("加密凭证失败: %w", err)
	}

	gitRepo := &model.GitRepo{
		Name:                name,
		URL:                 url,
		Provider:            provider,
		AuthType:            authType,
		CredentialEncrypted: encrypted,
	}

	if err := s.repo.Create(gitRepo); err != nil {
		return nil, fmt.Errorf("创建 Git 仓库失败: %w", err)
	}
	return gitRepo, nil
}

// GetByID 根据 ID 获取 Git 仓库
func (s *GitRepoService) GetByID(id uint) (*model.GitRepo, error) {
	return s.repo.FindByID(id)
}

// List 列出 Git 仓库
func (s *GitRepoService) List(page, pageSize int) ([]model.GitRepo, int64, error) {
	return s.repo.List(page, pageSize)
}

// Update 更新 Git 仓库
func (s *GitRepoService) Update(id uint, name, url string, credential *string) (*model.GitRepo, error) {
	gitRepo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("Git 仓库不存在: %w", err)
	}

	if name != "" {
		gitRepo.Name = name
	}
	if url != "" {
		gitRepo.URL = url
	}
	if credential != nil && *credential != "" {
		encrypted, err := s.cryptoSvc.Encrypt(*credential)
		if err != nil {
			return nil, fmt.Errorf("加密凭证失败: %w", err)
		}
		gitRepo.CredentialEncrypted = encrypted
	}

	if err := s.repo.Update(gitRepo); err != nil {
		return nil, fmt.Errorf("更新 Git 仓库失败: %w", err)
	}
	return gitRepo, nil
}

// Delete 删除 Git 仓库
func (s *GitRepoService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// GetDecryptedCredential 获取解密后的凭证
func (s *GitRepoService) GetDecryptedCredential(id uint) (string, error) {
	repo, err := s.repo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("Git 仓库不存在: %w", err)
	}
	return s.cryptoSvc.Decrypt(repo.CredentialEncrypted)
}

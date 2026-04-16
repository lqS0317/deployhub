package registry

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"

	"gorm.io/gorm"
)

// RegistryService 镜像仓库管理服务
type RegistryService struct {
	repo      repository.RegistryRepository
	cryptoSvc *crypto.CryptoService
}

// NewRegistryService 创建镜像仓库服务
func NewRegistryService(repo repository.RegistryRepository, cryptoSvc *crypto.CryptoService) *RegistryService {
	return &RegistryService{repo: repo, cryptoSvc: cryptoSvc}
}

// Create 创建镜像仓库（认证信息加密存储）
func (s *RegistryService) Create(name, url, provider, authConfig string) (*model.Registry, error) {
	if _, err := s.repo.FindByName(name); err == nil {
		return nil, fmt.Errorf("镜像仓库 %s 已存在", name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询镜像仓库失败: %w", err)
	}

	encrypted, err := s.cryptoSvc.Encrypt(authConfig)
	if err != nil {
		return nil, fmt.Errorf("加密认证信息失败: %w", err)
	}

	reg := &model.Registry{
		Name:                name,
		URL:                 url,
		Provider:            provider,
		AuthConfigEncrypted: encrypted,
	}

	if err := s.repo.Create(reg); err != nil {
		return nil, fmt.Errorf("创建镜像仓库失败: %w", err)
	}
	return reg, nil
}

// GetByID 根据 ID 获取
func (s *RegistryService) GetByID(id uint) (*model.Registry, error) {
	return s.repo.FindByID(id)
}

// List 列出镜像仓库
func (s *RegistryService) List(page, pageSize int) ([]model.Registry, int64, error) {
	return s.repo.List(page, pageSize)
}

// Update 更新镜像仓库
func (s *RegistryService) Update(id uint, name, url string, authConfig *string) (*model.Registry, error) {
	reg, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("镜像仓库不存在: %w", err)
	}

	if name != "" {
		reg.Name = name
	}
	if url != "" {
		reg.URL = url
	}
	if authConfig != nil && *authConfig != "" {
		encrypted, err := s.cryptoSvc.Encrypt(*authConfig)
		if err != nil {
			return nil, fmt.Errorf("加密认证信息失败: %w", err)
		}
		reg.AuthConfigEncrypted = encrypted
	}

	if err := s.repo.Update(reg); err != nil {
		return nil, fmt.Errorf("更新镜像仓库失败: %w", err)
	}
	return reg, nil
}

// Delete 删除镜像仓库
func (s *RegistryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

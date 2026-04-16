package cluster

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"

	"gorm.io/gorm"
)

// ClusterService 集群管理服务
type ClusterService struct {
	repo      repository.ClusterRepository
	cryptoSvc *crypto.CryptoService
}

// NewClusterService 创建集群服务
func NewClusterService(repo repository.ClusterRepository, cryptoSvc *crypto.CryptoService) *ClusterService {
	return &ClusterService{repo: repo, cryptoSvc: cryptoSvc}
}

// Create 创建集群（kubeconfig 加密存储）
func (s *ClusterService) Create(name, displayName, env, apiServer, kubeconfig, helmServiceAccount string) (*model.Cluster, error) {
	if _, err := s.repo.FindByName(name); err == nil {
		return nil, fmt.Errorf("集群 %s 已存在", name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询集群失败: %w", err)
	}

	encrypted, err := s.cryptoSvc.Encrypt(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("加密 kubeconfig 失败: %w", err)
	}

	cluster := &model.Cluster{
		Name:                name,
		DisplayName:         displayName,
		Env:                 env,
		APIServer:           apiServer,
		KubeconfigEncrypted: encrypted,
		Status:              "active",
		HelmServiceAccount:  helmServiceAccount,
	}

	if err := s.repo.Create(cluster); err != nil {
		return nil, fmt.Errorf("创建集群失败: %w", err)
	}

	return cluster, nil
}

// GetByID 根据 ID 获取集群
func (s *ClusterService) GetByID(id uint) (*model.Cluster, error) {
	return s.repo.FindByID(id)
}

// List 列出集群
func (s *ClusterService) List(page, pageSize int) ([]model.Cluster, int64, error) {
	return s.repo.List(page, pageSize)
}

// Update 更新集群
func (s *ClusterService) Update(id uint, displayName, env, apiServer string, kubeconfig, helmServiceAccount *string) (*model.Cluster, error) {
	cluster, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("集群不存在: %w", err)
	}

	if displayName != "" {
		cluster.DisplayName = displayName
	}
	if env != "" {
		cluster.Env = env
	}
	if apiServer != "" {
		cluster.APIServer = apiServer
	}
	if kubeconfig != nil && *kubeconfig != "" {
		encrypted, err := s.cryptoSvc.Encrypt(*kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("加密 kubeconfig 失败: %w", err)
		}
		cluster.KubeconfigEncrypted = encrypted
	}
	if helmServiceAccount != nil {
		cluster.HelmServiceAccount = *helmServiceAccount
	}

	if err := s.repo.Update(cluster); err != nil {
		return nil, fmt.Errorf("更新集群失败: %w", err)
	}
	return cluster, nil
}

// Delete 删除集群
func (s *ClusterService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// GetDecryptedKubeconfig 获取解密后的 kubeconfig
func (s *ClusterService) GetDecryptedKubeconfig(id uint) (string, error) {
	cluster, err := s.repo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("集群不存在: %w", err)
	}
	return s.cryptoSvc.Decrypt(cluster.KubeconfigEncrypted)
}

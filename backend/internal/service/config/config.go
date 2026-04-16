package config

import (
	"encoding/json"
	"fmt"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/cluster"
	"deployhub/internal/service/crypto"
)

// EnvValueSummary 环境变量摘要（脱敏展示）
type EnvValueSummary struct {
	ID        uint   `json:"id"`
	ClusterID uint   `json:"cluster_id"`
	HasValues bool   `json:"has_values"`
	UpdatedAt string `json:"updated_at"`
}

// ConfigService 配置中心服务
type ConfigService struct {
	templateRepo repository.ConfigTemplateRepository
	envValueRepo repository.ConfigEnvValueRepository
	versionRepo  repository.ConfigVersionRepository
	deployRepo   repository.ConfigDeploymentRepository
	cryptoSvc    *crypto.CryptoService
	clientPool   *cluster.ClientsetPool
}

// NewConfigService 创建配置中心服务实例
func NewConfigService(
	templateRepo repository.ConfigTemplateRepository,
	envValueRepo repository.ConfigEnvValueRepository,
	versionRepo repository.ConfigVersionRepository,
	deployRepo repository.ConfigDeploymentRepository,
	cryptoSvc *crypto.CryptoService,
	clientPool *cluster.ClientsetPool,
) *ConfigService {
	return &ConfigService{
		templateRepo: templateRepo,
		envValueRepo: envValueRepo,
		versionRepo:  versionRepo,
		deployRepo:   deployRepo,
		cryptoSvc:    cryptoSvc,
		clientPool:   clientPool,
	}
}

// CreateTemplate 创建配置模板
func (s *ConfigService) CreateTemplate(serviceID uint, name, configType, content string) (*model.ConfigTemplate, error) {
	if configType != "configmap" && configType != "secret" {
		return nil, fmt.Errorf("配置类型必须为 configmap 或 secret")
	}

	tpl := &model.ConfigTemplate{
		ServiceID:       serviceID,
		Name:            name,
		ConfigType:      configType,
		TemplateContent: content,
	}

	if err := s.templateRepo.Create(tpl); err != nil {
		return nil, fmt.Errorf("创建配置模板失败: %w", err)
	}
	return tpl, nil
}

// GetTemplate 获取配置模板详情
func (s *ConfigService) GetTemplate(id uint) (*model.ConfigTemplate, error) {
	tpl, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("配置模板不存在: %w", err)
	}
	return tpl, nil
}

// UpdateTemplate 更新配置模板名称和内容
func (s *ConfigService) UpdateTemplate(id uint, name, content string) (*model.ConfigTemplate, error) {
	tpl, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("配置模板不存在: %w", err)
	}

	tpl.Name = name
	tpl.TemplateContent = content

	if err := s.templateRepo.Update(tpl); err != nil {
		return nil, fmt.Errorf("更新配置模板失败: %w", err)
	}
	return tpl, nil
}

// DeleteTemplate 删除配置模板
func (s *ConfigService) DeleteTemplate(id uint) error {
	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("删除配置模板失败: %w", err)
	}
	return nil
}

// ListTemplates 列出指定服务的所有配置模板
func (s *ConfigService) ListTemplates(serviceID uint) ([]model.ConfigTemplate, error) {
	return s.templateRepo.ListByService(serviceID)
}

// SetEnvValues 设置配置环境变量值（AES-256-GCM 加密存储）
func (s *ConfigService) SetEnvValues(templateID, clusterID uint, vars map[string]string) error {
	jsonBytes, err := json.Marshal(vars)
	if err != nil {
		return fmt.Errorf("序列化环境变量失败: %w", err)
	}

	encrypted, err := s.cryptoSvc.Encrypt(string(jsonBytes))
	if err != nil {
		return fmt.Errorf("加密环境变量失败: %w", err)
	}

	val := &model.ConfigEnvValue{
		ConfigTemplateID: templateID,
		ClusterID:        clusterID,
		ValuesEncrypted:  encrypted,
	}

	if err := s.envValueRepo.CreateOrUpdate(val); err != nil {
		return fmt.Errorf("保存环境变量失败: %w", err)
	}
	return nil
}

// GetEnvValues 获取配置环境变量值（解密后返回明文 map）
func (s *ConfigService) GetEnvValues(templateID, clusterID uint) (map[string]string, error) {
	val, err := s.envValueRepo.FindByTemplateAndCluster(templateID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("环境变量不存在: %w", err)
	}

	plaintext, err := s.cryptoSvc.Decrypt(val.ValuesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("解密环境变量失败: %w", err)
	}

	var vars map[string]string
	if err := json.Unmarshal([]byte(plaintext), &vars); err != nil {
		return nil, fmt.Errorf("反序列化环境变量失败: %w", err)
	}

	return vars, nil
}

// ListEnvValues 列出模板下所有集群的环境变量摘要（脱敏）
func (s *ConfigService) ListEnvValues(templateID uint) ([]EnvValueSummary, error) {
	list, err := s.envValueRepo.ListByTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("查询环境变量列表失败: %w", err)
	}

	summaries := make([]EnvValueSummary, 0, len(list))
	for _, v := range list {
		summaries = append(summaries, EnvValueSummary{
			ID:        v.ID,
			ClusterID: v.ClusterID,
			HasValues: v.ValuesEncrypted != "",
			UpdatedAt: v.UpdatedAt.Format(time.RFC3339),
		})
	}
	return summaries, nil
}

// RenderPreview 预览渲染结果（不创建版本）
func (s *ConfigService) RenderPreview(templateID, clusterID uint) (string, error) {
	tpl, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return "", fmt.Errorf("配置模板不存在: %w", err)
	}

	vars, err := s.GetEnvValues(templateID, clusterID)
	if err != nil {
		return "", fmt.Errorf("获取环境变量失败: %w", err)
	}

	rendered, err := RenderTemplate(tpl.TemplateContent, vars)
	if err != nil {
		return "", fmt.Errorf("渲染配置失败: %w", err)
	}

	return rendered, nil
}

// CreateVersion 创建配置版本（自动递增版本号）
func (s *ConfigService) CreateVersion(templateID, clusterID, userID uint) (*model.ConfigVersion, error) {
	rendered, err := s.RenderPreview(templateID, clusterID)
	if err != nil {
		return nil, err
	}

	maxVer, err := s.versionRepo.GetMaxVersion(templateID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取最大版本号失败: %w", err)
	}

	ver := &model.ConfigVersion{
		ConfigTemplateID: templateID,
		ClusterID:        clusterID,
		Version:          maxVer + 1,
		RenderedContent:  rendered,
		CreatedByID:      userID,
	}

	if err := s.versionRepo.Create(ver); err != nil {
		return nil, fmt.Errorf("创建配置版本失败: %w", err)
	}

	return ver, nil
}

// GetVersion 获取指定版本详情
func (s *ConfigService) GetVersion(id uint) (*model.ConfigVersion, error) {
	ver, err := s.versionRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("配置版本不存在: %w", err)
	}
	return ver, nil
}

// ListVersions 列出指定模板+集群的所有版本
func (s *ConfigService) ListVersions(templateID, clusterID uint) ([]model.ConfigVersion, error) {
	return s.versionRepo.ListByTemplateAndCluster(templateID, clusterID)
}

// DeployConfig 将配置版本下发到 Kubernetes 集群
func (s *ConfigService) DeployConfig(versionID, clusterID, userID uint, namespace string) (*model.ConfigDeployment, error) {
	ver, err := s.versionRepo.FindByID(versionID)
	if err != nil {
		return nil, fmt.Errorf("配置版本不存在: %w", err)
	}

	tpl, err := s.templateRepo.FindByID(ver.ConfigTemplateID)
	if err != nil {
		return nil, fmt.Errorf("配置模板不存在: %w", err)
	}

	resourceName := fmt.Sprintf("%s-v%d", tpl.Name, ver.Version)

	dep := &model.ConfigDeployment{
		ConfigVersionID: versionID,
		ClusterID:       clusterID,
		Namespace:       namespace,
		ResourceName:    resourceName,
		Status:          "pending",
		DeployedByID:    userID,
	}

	if err := s.deployRepo.Create(dep); err != nil {
		return nil, fmt.Errorf("创建下发记录失败: %w", err)
	}

	clientset, err := s.clientPool.GetClientset(clusterID)
	if err != nil {
		dep.Status = "failed"
		s.deployRepo.Update(dep)
		return dep, fmt.Errorf("获取集群连接失败: %w", err)
	}

	if err := s.syncToK8s(clientset, namespace, resourceName, tpl.ConfigType, ver.RenderedContent); err != nil {
		dep.Status = "failed"
		s.deployRepo.Update(dep)
		return dep, fmt.Errorf("同步配置到 K8s 失败: %w", err)
	}

	now := time.Now()
	dep.Status = "success"
	dep.DeployedAt = &now
	if err := s.deployRepo.Update(dep); err != nil {
		return dep, fmt.Errorf("更新下发记录失败: %w", err)
	}

	return dep, nil
}

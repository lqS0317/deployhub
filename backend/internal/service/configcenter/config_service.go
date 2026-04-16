package configcenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ConfigService 配置中心核心服务
type ConfigService struct {
	entryRepo   repository.ConfigEntryRepository
	itemRepo    repository.ConfigItemRepository
	releaseRepo repository.ConfigReleaseRepository
	cryptoSvc   *crypto.CryptoService
}

// NewConfigService 创建配置中心服务
func NewConfigService(
	entryRepo repository.ConfigEntryRepository,
	itemRepo repository.ConfigItemRepository,
	releaseRepo repository.ConfigReleaseRepository,
	cryptoSvc *crypto.CryptoService,
) *ConfigService {
	return &ConfigService{
		entryRepo:   entryRepo,
		itemRepo:    itemRepo,
		releaseRepo: releaseRepo,
		cryptoSvc:   cryptoSvc,
	}
}

// PublishedConfig 已发布配置的聚合视图
type PublishedConfig struct {
	ConfigType string          `json:"config_type"`
	Snapshot   json.RawMessage `json:"snapshot"`
}

// ---- Entry 操作 ----

// CreateEntry 创建配置条目
func (s *ConfigService) CreateEntry(serviceID, clusterID uint, name, configType, format, mountPath string) (*model.ConfigEntry, error) {
	if !isValidConfigType(configType) {
		return nil, fmt.Errorf("不支持的配置类型: %s（仅支持 env/configmap/secret/serviceaccount）", configType)
	}
	if !isValidFormat(format) {
		return nil, fmt.Errorf("不支持的格式: %s（仅支持 properties/yaml/json）", format)
	}
	// env 和 serviceaccount 类型强制 properties 格式
	if configType == model.ConfigTypeEnv || configType == model.ConfigTypeServiceAccount {
		format = model.ConfigFormatProperties
	}
	// serviceaccount 类型每个服务每个环境只允许一个
	if configType == model.ConfigTypeServiceAccount {
		entries, _ := s.entryRepo.List(serviceID, clusterID)
		for _, e := range entries {
			if e.ConfigType == model.ConfigTypeServiceAccount {
				return nil, fmt.Errorf("该服务在此环境已有 ServiceAccount 配置条目: %s", e.Name)
			}
		}
	}

	existing, err := s.entryRepo.FindByName(serviceID, clusterID, name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("同名配置条目已存在: %s", name)
	}

	entry := &model.ConfigEntry{
		ServiceID:  serviceID,
		ClusterID:  clusterID,
		Name:       name,
		ConfigType: configType,
		Format:     format,
		MountPath:  mountPath,
	}
	if err := s.entryRepo.Create(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// UpdateEntry 更新配置条目
func (s *ConfigService) UpdateEntry(id uint, name, draftContent string) (*model.ConfigEntry, error) {
	entry, err := s.entryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}
	if name != "" {
		entry.Name = name
	}
	entry.DraftContent = draftContent
	if err := s.entryRepo.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// DeleteEntry 删除配置条目
func (s *ConfigService) DeleteEntry(id uint) error {
	return s.entryRepo.Delete(id)
}

// ListEntries 列出配置条目
func (s *ConfigService) ListEntries(serviceID, clusterID uint) ([]model.ConfigEntry, error) {
	return s.entryRepo.List(serviceID, clusterID)
}

// GetEntry 获取配置条目
func (s *ConfigService) GetEntry(id uint) (*model.ConfigEntry, error) {
	entry, err := s.entryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}
	return entry, nil
}

// ---- Item 操作 ----

// ListItems 列出配置项（secret 类型不解密时掩码值）
func (s *ConfigService) ListItems(entryID uint, decrypt bool) ([]model.ConfigItem, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}

	items, err := s.itemRepo.List(entryID)
	if err != nil {
		return nil, err
	}

	if entry.ConfigType == model.ConfigTypeSecret {
		for i := range items {
			if decrypt {
				plain, decErr := s.cryptoSvc.Decrypt(items[i].Value)
				if decErr != nil {
					return nil, fmt.Errorf("解密配置项 %s 失败: %w", items[i].Key, decErr)
				}
				items[i].Value = plain
			} else {
				items[i].Value = "******"
			}
		}
	}
	return items, nil
}

// CreateItem 创建配置项（仅 properties 格式）
func (s *ConfigService) CreateItem(entryID uint, key, value, comment string) (*model.ConfigItem, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}
	if entry.Format != model.ConfigFormatProperties {
		return nil, fmt.Errorf("仅 properties 格式支持逐项管理")
	}

	_, err = s.itemRepo.FindByKey(entryID, key)
	if err == nil {
		return nil, fmt.Errorf("配置项 key=%s 已存在", key)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if entry.ConfigType == model.ConfigTypeSecret {
		encrypted, encErr := s.cryptoSvc.Encrypt(value)
		if encErr != nil {
			return nil, fmt.Errorf("加密失败: %w", encErr)
		}
		value = encrypted
	}

	item := &model.ConfigItem{
		ConfigEntryID: entryID,
		Key:           key,
		Value:         value,
		Comment:       comment,
	}
	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateItem 更新配置项
func (s *ConfigService) UpdateItem(id uint, value, comment string) (*model.ConfigItem, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	entry, err := s.entryRepo.FindByID(item.ConfigEntryID)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}

	if entry.ConfigType == model.ConfigTypeSecret {
		encrypted, encErr := s.cryptoSvc.Encrypt(value)
		if encErr != nil {
			return nil, fmt.Errorf("加密失败: %w", encErr)
		}
		value = encrypted
	}

	item.Value = value
	item.Comment = comment
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteItem 软删除配置项
func (s *ConfigService) DeleteItem(id uint) error {
	return s.itemRepo.SoftDelete(id)
}

// ---- Draft 操作 ----

// GetDraft 获取草稿内容
func (s *ConfigService) GetDraft(entryID uint) (string, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return "", fmt.Errorf("配置条目不存在")
	}

	if entry.Format == model.ConfigFormatProperties {
		items, itemErr := s.itemRepo.List(entryID)
		if itemErr != nil {
			return "", itemErr
		}
		return buildPropertiesString(items), nil
	}

	return entry.DraftContent, nil
}

// SaveDraft 保存草稿（仅 yaml/json 格式）
func (s *ConfigService) SaveDraft(entryID uint, content string) error {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return fmt.Errorf("配置条目不存在")
	}
	if entry.Format == model.ConfigFormatProperties {
		return fmt.Errorf("properties 格式不支持直接编辑草稿，请使用配置项管理")
	}

	entry.DraftContent = content
	return s.entryRepo.Update(entry)
}

// ---- Publish / Rollback ----

// Publish 发布配置
func (s *ConfigService) Publish(entryID, userID uint, comment string) (*model.ConfigRelease, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}

	snapshot, err := s.buildSnapshot(entry, entryID)
	if err != nil {
		return nil, err
	}

	nextVer, err := s.releaseRepo.GetNextVersion(entryID)
	if err != nil {
		return nil, err
	}

	release := &model.ConfigRelease{
		ConfigEntryID: entryID,
		Version:       nextVer,
		Snapshot:      snapshot,
		Status:        model.ReleaseStatusPublished,
		Comment:       comment,
		CreatedByID:   userID,
	}
	if err := s.releaseRepo.Create(release); err != nil {
		return nil, err
	}

	if entry.Format == model.ConfigFormatProperties {
		if err := s.itemRepo.PurgeDeleted(entryID); err != nil {
			return nil, err
		}
	}

	return release, nil
}

// Rollback 回滚到指定版本
func (s *ConfigService) Rollback(entryID, userID uint, targetVersion int, comment string) (*model.ConfigRelease, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("配置条目不存在")
	}

	releases, err := s.releaseRepo.List(entryID)
	if err != nil {
		return nil, err
	}
	var targetRelease *model.ConfigRelease
	for i := range releases {
		if releases[i].Version == targetVersion {
			targetRelease = &releases[i]
			break
		}
	}
	if targetRelease == nil {
		return nil, fmt.Errorf("未找到版本 %d 的发布记录", targetVersion)
	}

	if err := s.restoreSnapshot(entry, entryID, targetRelease.Snapshot); err != nil {
		return nil, fmt.Errorf("恢复快照失败: %w", err)
	}

	nextVer, err := s.releaseRepo.GetNextVersion(entryID)
	if err != nil {
		return nil, err
	}

	newRelease := &model.ConfigRelease{
		ConfigEntryID: entryID,
		Version:       nextVer,
		Snapshot:      targetRelease.Snapshot,
		Status:        model.ReleaseStatusRolledBack,
		Comment:       fmt.Sprintf("回滚至 v%d: %s", targetVersion, comment),
		CreatedByID:   userID,
	}
	if err := s.releaseRepo.Create(newRelease); err != nil {
		return nil, err
	}
	return newRelease, nil
}

// ListReleases 列出发布历史
func (s *ConfigService) ListReleases(entryID uint) ([]model.ConfigRelease, error) {
	return s.releaseRepo.List(entryID)
}

// ---- 聚合查询 ----

// GetPublishedSnapshot 获取指定配置条目的最新发布快照
func (s *ConfigService) GetPublishedSnapshot(entryID uint) (*model.ConfigRelease, error) {
	release, err := s.releaseRepo.FindLatestPublished(entryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return release, nil
}

// ---- 内部方法 ----

// buildSnapshot 构建发布快照
func (s *ConfigService) buildSnapshot(entry *model.ConfigEntry, entryID uint) (datatypes.JSON, error) {
	if entry.Format == model.ConfigFormatProperties {
		items, err := s.itemRepo.List(entryID)
		if err != nil {
			return nil, err
		}
		type itemSnapshot struct {
			Key     string `json:"key"`
			Value   string `json:"value"`
			Comment string `json:"comment"`
		}
		snapItems := make([]itemSnapshot, 0, len(items))
		for _, it := range items {
			snapItems = append(snapItems, itemSnapshot{
				Key:     it.Key,
				Value:   it.Value,
				Comment: it.Comment,
			})
		}
		data, err := json.Marshal(snapItems)
		return datatypes.JSON(data), err
	}

	data, err := json.Marshal(entry.DraftContent)
	return datatypes.JSON(data), err
}

// restoreSnapshot 从快照恢复配置
func (s *ConfigService) restoreSnapshot(entry *model.ConfigEntry, entryID uint, snapshot datatypes.JSON) error {
	if entry.Format == model.ConfigFormatProperties {
		type itemSnapshot struct {
			Key     string `json:"key"`
			Value   string `json:"value"`
			Comment string `json:"comment"`
		}
		var items []itemSnapshot
		if err := json.Unmarshal(snapshot, &items); err != nil {
			return err
		}

		allItems, err := s.itemRepo.ListAll(entryID)
		if err != nil {
			return err
		}
		for _, old := range allItems {
			if err := s.itemRepo.SoftDelete(old.ID); err != nil {
				return err
			}
		}
		if err := s.itemRepo.PurgeDeleted(entryID); err != nil {
			return err
		}

		for _, it := range items {
			newItem := &model.ConfigItem{
				ConfigEntryID: entryID,
				Key:           it.Key,
				Value:         it.Value,
				Comment:       it.Comment,
			}
			if err := s.itemRepo.Create(newItem); err != nil {
				return err
			}
		}
		return nil
	}

	var content string
	if err := json.Unmarshal(snapshot, &content); err != nil {
		return err
	}
	entry.DraftContent = content
	return s.entryRepo.Update(entry)
}

// buildPropertiesString 从配置项列表生成 KV 文本
func buildPropertiesString(items []model.ConfigItem) string {
	var sb strings.Builder
	for i, item := range items {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(item.Key)
		sb.WriteString("=")
		sb.WriteString(item.Value)
	}
	return sb.String()
}

func isValidFormat(f string) bool {
	return f == model.ConfigFormatProperties || f == model.ConfigFormatYAML || f == model.ConfigFormatJSON
}

func isValidConfigType(t string) bool {
	return t == model.ConfigTypeEnv || t == model.ConfigTypeConfigMap || t == model.ConfigTypeSecret || t == model.ConfigTypeServiceAccount
}

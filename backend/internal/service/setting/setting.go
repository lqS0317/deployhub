package setting

import (
	"strings"
	"sync"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// SettingService 系统配置服务，带内存缓存（TTL 30 秒）
type SettingService struct {
	repo  repository.SystemSettingRepository
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	value   string
	expires time.Time
}

const cacheTTL = 30 * time.Second

func NewSettingService(repo repository.SystemSettingRepository) *SettingService {
	return &SettingService{
		repo:  repo,
		cache: make(map[string]cacheEntry),
	}
}

// Get 获取单个配置值，优先从缓存读取
func (s *SettingService) Get(key string) string {
	s.mu.RLock()
	if e, ok := s.cache[key]; ok && time.Now().Before(e.expires) {
		s.mu.RUnlock()
		return e.value
	}
	s.mu.RUnlock()

	setting, err := s.repo.Get(key)
	if err != nil {
		return ""
	}

	s.mu.Lock()
	s.cache[key] = cacheEntry{value: setting.Value, expires: time.Now().Add(cacheTTL)}
	s.mu.Unlock()

	return setting.Value
}

// GetAll 获取所有配置项
func (s *SettingService) GetAll() ([]model.SystemSetting, error) {
	return s.repo.GetAll()
}

// Set 设置配置项并清除缓存
func (s *SettingService) Set(key, value, description string) error {
	setting := &model.SystemSetting{
		Key:         key,
		Value:       value,
		Description: description,
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Upsert(setting); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()

	return nil
}

// GetHelmJobNamespace 获取 Helm Job 命名空间
func (s *SettingService) GetHelmJobNamespace() string {
	v := s.Get(model.SettingHelmJobNamespace)
	if v == "" {
		return "deployhub-jobs"
	}
	return v
}

// GetEnvValuesMap 获取 env values 映射
func (s *SettingService) GetEnvValuesMap() map[string]string {
	raw := s.Get(model.SettingEnvValuesMap)
	return parseEnvValuesMap(raw)
}

func parseEnvValuesMap(raw string) map[string]string {
	m := make(map[string]string)
	if raw == "" {
		return m
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return m
}

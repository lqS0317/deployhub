package config

import (
	"encoding/hex"
	"fmt"
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/service/crypto"

	"github.com/stretchr/testify/assert"
)

// --- Mock Repositories ---

type mockTemplateRepo struct {
	templates map[uint]*model.ConfigTemplate
	nextID    uint
}

func newMockTemplateRepo() *mockTemplateRepo {
	return &mockTemplateRepo{templates: make(map[uint]*model.ConfigTemplate), nextID: 1}
}

func (m *mockTemplateRepo) Create(tpl *model.ConfigTemplate) error {
	tpl.ID = m.nextID
	m.nextID++
	m.templates[tpl.ID] = tpl
	return nil
}

func (m *mockTemplateRepo) FindByID(id uint) (*model.ConfigTemplate, error) {
	tpl, ok := m.templates[id]
	if !ok {
		return nil, assert.AnError
	}
	return tpl, nil
}

func (m *mockTemplateRepo) Update(tpl *model.ConfigTemplate) error {
	m.templates[tpl.ID] = tpl
	return nil
}

func (m *mockTemplateRepo) Delete(id uint) error {
	delete(m.templates, id)
	return nil
}

func (m *mockTemplateRepo) ListByService(serviceID uint) ([]model.ConfigTemplate, error) {
	var result []model.ConfigTemplate
	for _, tpl := range m.templates {
		if tpl.ServiceID == serviceID {
			result = append(result, *tpl)
		}
	}
	return result, nil
}

type mockEnvValueRepo struct {
	values map[string]*model.ConfigEnvValue // key: "templateID-clusterID"
	nextID uint
}

func newMockEnvValueRepo() *mockEnvValueRepo {
	return &mockEnvValueRepo{values: make(map[string]*model.ConfigEnvValue), nextID: 1}
}

func (m *mockEnvValueRepo) key(templateID, clusterID uint) string {
	return fmt.Sprintf("%d-%d", templateID, clusterID)
}

func (m *mockEnvValueRepo) CreateOrUpdate(val *model.ConfigEnvValue) error {
	k := m.key(val.ConfigTemplateID, val.ClusterID)
	if existing, ok := m.values[k]; ok {
		existing.ValuesEncrypted = val.ValuesEncrypted
		return nil
	}
	val.ID = m.nextID
	m.nextID++
	m.values[k] = val
	return nil
}

func (m *mockEnvValueRepo) FindByTemplateAndCluster(templateID, clusterID uint) (*model.ConfigEnvValue, error) {
	k := m.key(templateID, clusterID)
	val, ok := m.values[k]
	if !ok {
		return nil, assert.AnError
	}
	return val, nil
}

func (m *mockEnvValueRepo) ListByTemplate(templateID uint) ([]model.ConfigEnvValue, error) {
	var result []model.ConfigEnvValue
	for _, v := range m.values {
		if v.ConfigTemplateID == templateID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (m *mockEnvValueRepo) Delete(id uint) error {
	for k, v := range m.values {
		if v.ID == id {
			delete(m.values, k)
			break
		}
	}
	return nil
}

type mockVersionRepo struct {
	versions map[uint]*model.ConfigVersion
	nextID   uint
}

func newMockVersionRepo() *mockVersionRepo {
	return &mockVersionRepo{versions: make(map[uint]*model.ConfigVersion), nextID: 1}
}

func (m *mockVersionRepo) Create(ver *model.ConfigVersion) error {
	ver.ID = m.nextID
	m.nextID++
	m.versions[ver.ID] = ver
	return nil
}

func (m *mockVersionRepo) FindByID(id uint) (*model.ConfigVersion, error) {
	ver, ok := m.versions[id]
	if !ok {
		return nil, assert.AnError
	}
	return ver, nil
}

func (m *mockVersionRepo) ListByTemplateAndCluster(templateID, clusterID uint) ([]model.ConfigVersion, error) {
	var result []model.ConfigVersion
	for _, v := range m.versions {
		if v.ConfigTemplateID == templateID && v.ClusterID == clusterID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (m *mockVersionRepo) GetMaxVersion(templateID, clusterID uint) (int, error) {
	maxVer := 0
	for _, v := range m.versions {
		if v.ConfigTemplateID == templateID && v.ClusterID == clusterID && v.Version > maxVer {
			maxVer = v.Version
		}
	}
	return maxVer, nil
}

type mockDeployRepo struct {
	deployments map[uint]*model.ConfigDeployment
	nextID      uint
}

func newMockDeployRepo() *mockDeployRepo {
	return &mockDeployRepo{deployments: make(map[uint]*model.ConfigDeployment), nextID: 1}
}

func (m *mockDeployRepo) Create(dep *model.ConfigDeployment) error {
	dep.ID = m.nextID
	m.nextID++
	m.deployments[dep.ID] = dep
	return nil
}

func (m *mockDeployRepo) FindByID(id uint) (*model.ConfigDeployment, error) {
	dep, ok := m.deployments[id]
	if !ok {
		return nil, assert.AnError
	}
	return dep, nil
}

func (m *mockDeployRepo) Update(dep *model.ConfigDeployment) error {
	m.deployments[dep.ID] = dep
	return nil
}

func (m *mockDeployRepo) ListByVersion(versionID uint) ([]model.ConfigDeployment, error) {
	var result []model.ConfigDeployment
	for _, d := range m.deployments {
		if d.ConfigVersionID == versionID {
			result = append(result, *d)
		}
	}
	return result, nil
}

// --- 辅助函数 ---

func newTestCryptoService(t *testing.T) *crypto.CryptoService {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	svc, err := crypto.NewCryptoService(hex.EncodeToString(key))
	assert.NoError(t, err)
	return svc
}

func newTestConfigService(t *testing.T) (*ConfigService, *mockTemplateRepo, *mockEnvValueRepo, *mockVersionRepo) {
	t.Helper()
	tplRepo := newMockTemplateRepo()
	envRepo := newMockEnvValueRepo()
	verRepo := newMockVersionRepo()
	depRepo := newMockDeployRepo()
	cryptoSvc := newTestCryptoService(t)

	svc := NewConfigService(tplRepo, envRepo, verRepo, depRepo, cryptoSvc, nil)
	return svc, tplRepo, envRepo, verRepo
}

// --- 模板 CRUD 测试 ---

func TestCreateTemplate(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	tpl, err := svc.CreateTemplate(1, "nginx-config", "configmap", "server { listen {{.PORT}}; }")
	assert.NoError(t, err)
	assert.Equal(t, uint(1), tpl.ID)
	assert.Equal(t, "nginx-config", tpl.Name)
	assert.Equal(t, "configmap", tpl.ConfigType)
}

func TestCreateTemplate_InvalidType(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	_, err := svc.CreateTemplate(1, "test", "invalid", "content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configmap 或 secret")
}

func TestGetTemplate(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	created, _ := svc.CreateTemplate(1, "test", "configmap", "content")
	got, err := svc.GetTemplate(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestUpdateTemplate(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	created, _ := svc.CreateTemplate(1, "old-name", "configmap", "old content")
	updated, err := svc.UpdateTemplate(created.ID, "new-name", "new content")
	assert.NoError(t, err)
	assert.Equal(t, "new-name", updated.Name)
	assert.Equal(t, "new content", updated.TemplateContent)
}

func TestDeleteTemplate(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	created, _ := svc.CreateTemplate(1, "to-delete", "secret", "content")
	err := svc.DeleteTemplate(created.ID)
	assert.NoError(t, err)

	_, err = svc.GetTemplate(created.ID)
	assert.Error(t, err)
}

func TestListTemplates(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.CreateTemplate(1, "tpl-a", "configmap", "a")
	svc.CreateTemplate(1, "tpl-b", "secret", "b")
	svc.CreateTemplate(2, "tpl-c", "configmap", "c")

	list, err := svc.ListTemplates(1)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

// --- 环境变量加密/解密测试 ---

func TestSetAndGetEnvValues(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	vars := map[string]string{
		"DB_HOST": "postgres.svc",
		"DB_PORT": "5432",
	}

	err := svc.SetEnvValues(1, 1, vars)
	assert.NoError(t, err)

	got, err := svc.GetEnvValues(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "postgres.svc", got["DB_HOST"])
	assert.Equal(t, "5432", got["DB_PORT"])
}

func TestSetEnvValues_Overwrite(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.SetEnvValues(1, 1, map[string]string{"KEY": "old"})
	svc.SetEnvValues(1, 1, map[string]string{"KEY": "new"})

	got, err := svc.GetEnvValues(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "new", got["KEY"])
}

func TestListEnvValues(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.SetEnvValues(1, 1, map[string]string{"A": "1"})
	svc.SetEnvValues(1, 2, map[string]string{"B": "2"})

	summaries, err := svc.ListEnvValues(1)
	assert.NoError(t, err)
	assert.Len(t, summaries, 2)
	for _, s := range summaries {
		assert.True(t, s.HasValues)
	}
}

// --- 版本创建与自增测试 ---

func TestCreateVersion_AutoIncrement(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.CreateTemplate(1, "tpl", "configmap", "host: {{.HOST}}")
	svc.SetEnvValues(1, 1, map[string]string{"HOST": "localhost"})

	v1, err := svc.CreateVersion(1, 1, 100)
	assert.NoError(t, err)
	assert.Equal(t, 1, v1.Version)
	assert.Contains(t, v1.RenderedContent, "host: localhost")

	v2, err := svc.CreateVersion(1, 1, 100)
	assert.NoError(t, err)
	assert.Equal(t, 2, v2.Version)
}

func TestListVersions(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.CreateTemplate(1, "tpl", "configmap", "val: {{.V}}")
	svc.SetEnvValues(1, 1, map[string]string{"V": "1"})

	svc.CreateVersion(1, 1, 100)
	svc.CreateVersion(1, 1, 100)

	list, err := svc.ListVersions(1, 1)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

// --- 渲染预览测试 ---

func TestRenderPreview(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.CreateTemplate(1, "tpl", "configmap", "db: {{.DB_HOST}}:{{.DB_PORT}}")
	svc.SetEnvValues(1, 1, map[string]string{"DB_HOST": "pg.svc", "DB_PORT": "5432"})

	result, err := svc.RenderPreview(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "db: pg.svc:5432", result)
}

func TestRenderPreview_MissingEnvValues(t *testing.T) {
	svc, _, _, _ := newTestConfigService(t)

	svc.CreateTemplate(1, "tpl", "configmap", "val: {{.MISSING}}")

	_, err := svc.RenderPreview(1, 1)
	assert.Error(t, err)
}

// --- Diff 测试 ---

func TestDiffVersions(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline2-modified\nline3\nline4"

	diff := DiffVersions(old, new)
	assert.Contains(t, diff, "- line2")
	assert.Contains(t, diff, "+ line2-modified")
	assert.Contains(t, diff, "+ line4")
}

package deploy

import (
	"encoding/json"
	"fmt"
	"strings"

	"deployhub/internal/model"
	"deployhub/internal/service/configcenter"
	"deployhub/internal/service/crypto"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// ConfigMountInfo 描述单个配置条目在容器中的挂载信息
type ConfigMountInfo struct {
	EntryName  string // 配置条目名称
	ConfigType string // env, configmap, secret, serviceaccount
	K8sName    string // K8s 资源名称: {svc}-{entry_name}
	MountPath  string // 挂载路径（configmap/secret 用，env/sa 为空）
}

// ConfigDeployResult 部署配置生成结果
type ConfigDeployResult struct {
	YAML               string            // ConfigMap/Secret/SA 资源 YAML
	Mounts             []ConfigMountInfo // 挂载信息
	ServiceAccountName string            // 如果有 SA 类型条目，返回 SA 名称
}

// ConfigDeployHelper 从配置中心读取已发布配置，生成 K8s ConfigMap/Secret YAML
type ConfigDeployHelper struct {
	configSvc *configcenter.ConfigService
	cryptoSvc *crypto.CryptoService
}

func NewConfigDeployHelper(configSvc *configcenter.ConfigService, cryptoSvc *crypto.CryptoService) *ConfigDeployHelper {
	return &ConfigDeployHelper{configSvc: configSvc, cryptoSvc: cryptoSvc}
}

// GenerateConfigResources 自动查找该服务+集群下所有已发布的 ConfigEntry，生成 K8s 资源
func (h *ConfigDeployHelper) GenerateConfigResources(serviceID, clusterID uint, serviceName, deployType string) (*ConfigDeployResult, error) {
	result := &ConfigDeployResult{}
	if h.configSvc == nil || deployType == "helm" {
		return result, nil
	}

	entries, err := h.configSvc.ListEntries(serviceID, clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取配置条目列表失败: %w", err)
	}
	if len(entries) == 0 {
		return result, nil
	}

	var parts []string

	for _, entry := range entries {
		release, err := h.configSvc.GetPublishedSnapshot(entry.ID)
		if err != nil {
			return nil, fmt.Errorf("获取配置条目 %s 的发布快照失败: %w", entry.Name, err)
		}
		if release == nil {
			continue
		}

		k8sName := fmt.Sprintf("%s-%s", serviceName, entry.Name)

		switch entry.ConfigType {
		case model.ConfigTypeEnv:
			cmYAML, err := h.buildConfigMap(k8sName, json.RawMessage(release.Snapshot))
			if err != nil {
				return nil, fmt.Errorf("生成 env ConfigMap 失败: %w", err)
			}
			parts = append(parts, cmYAML)

		case model.ConfigTypeConfigMap:
			cmYAML, err := h.buildConfigMap(k8sName, json.RawMessage(release.Snapshot))
			if err != nil {
				return nil, fmt.Errorf("生成 ConfigMap 失败: %w", err)
			}
			parts = append(parts, cmYAML)

		case model.ConfigTypeSecret:
			secretYAML, err := h.buildSecret(k8sName, json.RawMessage(release.Snapshot))
			if err != nil {
				return nil, fmt.Errorf("生成 Secret 失败: %w", err)
			}
			parts = append(parts, secretYAML)

		case model.ConfigTypeServiceAccount:
			saName := fmt.Sprintf("%s-sa", serviceName)
			saYAML, err := h.buildServiceAccount(saName, json.RawMessage(release.Snapshot))
			if err != nil {
				return nil, fmt.Errorf("生成 ServiceAccount 失败: %w", err)
			}
			parts = append(parts, saYAML)
			result.ServiceAccountName = saName
		}

		// 挂载路径
		mountPath := entry.MountPath
		if mountPath == "" && entry.ConfigType != model.ConfigTypeEnv && entry.ConfigType != model.ConfigTypeServiceAccount {
			mountPath = fmt.Sprintf("/etc/config/%s", entry.Name)
		}

		result.Mounts = append(result.Mounts, ConfigMountInfo{
			EntryName:  entry.Name,
			ConfigType: entry.ConfigType,
			K8sName:    k8sName,
			MountPath:  mountPath,
		})
	}

	result.YAML = strings.Join(parts, "---\n")
	return result, nil
}

// buildServiceAccount 从快照 KV 生成 ServiceAccount YAML（KV 作为 annotations）
func (h *ConfigDeployHelper) buildServiceAccount(name string, snapshot json.RawMessage) (string, error) {
	annotations, err := h.parseSnapshotData(snapshot, false)
	if err != nil {
		return "", err
	}

	sa := &corev1.ServiceAccount{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations, Labels: map[string]string{"managed-by": "deployhub-config"}},
	}
	y, err := yaml.Marshal(sa)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// buildConfigMap 从快照生成 ConfigMap YAML
func (h *ConfigDeployHelper) buildConfigMap(name string, snapshot json.RawMessage) (string, error) {
	data, err := h.parseSnapshotData(snapshot, false)
	if err != nil {
		return "", err
	}

	cm := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"managed-by": "deployhub-config"}},
		Data:       data,
	}
	y, err := yaml.Marshal(cm)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// buildSecret 从快照生成 Secret YAML（值解密）
func (h *ConfigDeployHelper) buildSecret(name string, snapshot json.RawMessage) (string, error) {
	data, err := h.parseSnapshotData(snapshot, true)
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"managed-by": "deployhub-config"}},
		StringData: data,
		Type:       corev1.SecretTypeOpaque,
	}
	y, err := yaml.Marshal(secret)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

// parseSnapshotData 解析快照为 key-value map
func (h *ConfigDeployHelper) parseSnapshotData(snapshot json.RawMessage, decryptValues bool) (map[string]string, error) {
	// 尝试解析为 KV 数组 [{key, value, comment}]
	var items []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(snapshot, &items); err == nil && len(items) > 0 {
		data := make(map[string]string, len(items))
		for _, item := range items {
			val := item.Value
			if decryptValues && h.cryptoSvc != nil {
				if decrypted, err := h.cryptoSvc.Decrypt(val); err == nil {
					val = decrypted
				}
			}
			data[item.Key] = val
		}
		return data, nil
	}

	// 尝试解析为字符串（yaml/json 格式的整体内容）
	var content string
	if err := json.Unmarshal(snapshot, &content); err == nil && content != "" {
		return map[string]string{"config": content}, nil
	}

	// 尝试解析为 map
	var kvMap map[string]string
	if err := json.Unmarshal(snapshot, &kvMap); err == nil {
		if decryptValues && h.cryptoSvc != nil {
			for k, v := range kvMap {
				if decrypted, err := h.cryptoSvc.Decrypt(v); err == nil {
					kvMap[k] = decrypted
				}
			}
		}
		return kvMap, nil
	}

	return map[string]string{}, nil
}

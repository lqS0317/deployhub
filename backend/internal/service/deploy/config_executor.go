package deploy

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/service/cluster"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// ConfigExecutor 根据表单配置生成 K8s YAML 后通过 YamlExecutor 应用
type ConfigExecutor struct {
	clientPool       *cluster.ClientsetPool
	yamlExecutor     *YamlExecutor
	configDeployHelper *ConfigDeployHelper
}

func NewConfigExecutor(clientPool *cluster.ClientsetPool) *ConfigExecutor {
	return &ConfigExecutor{
		clientPool:   clientPool,
		yamlExecutor: NewYamlExecutor(clientPool),
	}
}

// SetConfigDeployHelper 注入配置中心部署助手（可选，避免循环依赖）
func (e *ConfigExecutor) SetConfigDeployHelper(h *ConfigDeployHelper) {
	e.configDeployHelper = h
}

// Execute 生成 YAML 并应用
func (e *ConfigExecutor) Execute(deployment *model.Deployment, service *model.Service) error {
	yamlStr, err := e.generateYAML(deployment, service)
	if err != nil {
		return fmt.Errorf("生成 YAML 失败: %w", err)
	}

	deployment.DeployCommand = "# ConfigExecutor 生成的 YAML\n" + yamlStr
	return e.yamlExecutor.ExecuteRaw(yamlStr, deployment, service)
}

// DryRun 生成 YAML 并做 dry-run 校验
func (e *ConfigExecutor) DryRun(deployment *model.Deployment, service *model.Service) (string, error) {
	yamlStr, err := e.generateYAML(deployment, service)
	if err != nil {
		return "", fmt.Errorf("生成 YAML 失败: %w", err)
	}
	return e.yamlExecutor.DryRunRaw(yamlStr, deployment, service)
}

// --- YAML 生成 ---

func (e *ConfigExecutor) generateYAML(dep *model.Deployment, svc *model.Service) (string, error) {
	var fullImage string
	if dep.ExternalImage != "" {
		fullImage = dep.ExternalImage
	} else {
		imageRepo := resolveImageRepo(dep, svc)
		imageTag := dep.ImageTag
		if imageTag == "" {
			imageTag = "latest"
		}
		fullImage = imageTag
		if imageRepo != "" && !strings.Contains(imageTag, "/") {
			fullImage = imageRepo + ":" + imageTag
		}
	}

	port := resolvePort(dep, svc)
	replicasVal := svc.DefaultReplicas
	if replicasVal <= 0 {
		replicasVal = svc.Replicas
	}
	if replicasVal <= 0 {
		replicasVal = 1
	}
	replicas := int32(replicasVal)
	name := svc.Name
	labels := map[string]string{"app": name, "managed-by": "deployhub"}

	command := parseStringList(json.RawMessage(svc.DefaultCommand))
	args := parseStringList(json.RawMessage(svc.DefaultArgs))
	liveness := parseProbe(json.RawMessage(svc.DefaultLivenessProbe))
	readiness := parseProbe(json.RawMessage(svc.DefaultReadinessProbe))

	// 从配置中心自动生成 ConfigMap/Secret 并获取挂载信息
	var configYAMLParts []string
	var envFromSources []corev1.EnvFromSource
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	// 从配置中心生成配置资源
	var serviceAccountName string
	if e.configDeployHelper != nil {
		configResult, err := e.configDeployHelper.GenerateConfigResources(svc.ID, dep.ClusterID, svc.Name, svc.DeployType)
		if err != nil {
			log.Printf("[ConfigExecutor] 生成配置资源失败（非致命）: %v", err)
		} else if configResult != nil {
			if configResult.YAML != "" {
				configYAMLParts = append(configYAMLParts, configResult.YAML)
			}
			serviceAccountName = configResult.ServiceAccountName

			for _, mi := range configResult.Mounts {
				switch mi.ConfigType {
				case "env":
					envFromSources = append(envFromSources, corev1.EnvFromSource{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: mi.K8sName},
						},
					})
				case "configmap":
					volName := "cfg-" + mi.EntryName
					volumes = append(volumes, corev1.Volume{
						Name: volName,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: mi.K8sName},
							},
						},
					})
					volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: volName, MountPath: mi.MountPath, ReadOnly: true})
				case "secret":
					volName := "sec-" + mi.EntryName
					volumes = append(volumes, corev1.Volume{
						Name: volName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{SecretName: mi.K8sName},
						},
					})
					volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: volName, MountPath: mi.MountPath, ReadOnly: true})
				}
			}
		}
	}

	container := corev1.Container{
		Name:            name,
		Image:           fullImage,
		ImagePullPolicy: corev1.PullAlways,
		Command:         command,
		Args:            args,
		EnvFrom:         envFromSources,
		VolumeMounts:    volumeMounts,
		LivenessProbe:   liveness,
		ReadinessProbe:  readiness,
	}

	if port > 0 {
		container.Ports = []corev1.ContainerPort{{Name: "main", ContainerPort: int32(port), Protocol: corev1.ProtocolTCP}}
	}

	cpuReq := svc.DefaultCPURequest
	if cpuReq == "" { cpuReq = svc.CPURequest }
	memReq := svc.DefaultMemRequest
	if memReq == "" { memReq = svc.MemRequest }
	cpuLim := svc.DefaultCPULimit
	if cpuLim == "" { cpuLim = svc.CPULimit }
	memLim := svc.DefaultMemLimit
	if memLim == "" { memLim = svc.MemLimit }
	container.Resources = buildResources(cpuReq, memReq, cpuLim, memLim)

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
		Volumes:    volumes,
	}
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
	}

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
			Annotations: map[string]string{
				"deployhub.io/deploy-timestamp": strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
		Spec: podSpec,
	}

	var parts []string
	// 先放配置中心生成的 ConfigMap/Secret
	parts = append(parts, configYAMLParts...)

	wt := resolveWorkloadType(dep, svc)

	if wt == "statefulset" {
		sts := &appsv1.StatefulSet{
			TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: dep.Namespace, Labels: labels},
			Spec: appsv1.StatefulSetSpec{
				Replicas:    &replicas,
				ServiceName: name,
				Selector:    &metav1.LabelSelector{MatchLabels: labels},
				Template:    podTemplate,
			},
		}
		y, _ := yaml.Marshal(sts)
		parts = append(parts, string(y))
	} else {
		k8sDep := &appsv1.Deployment{
			TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: dep.Namespace, Labels: labels},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{MatchLabels: labels},
				Template: podTemplate,
			},
		}
		y, _ := yaml.Marshal(k8sDep)
		parts = append(parts, string(y))
	}

	result := strings.Join(parts, "---\n")
	log.Printf("[ConfigExecutor] 为服务 %s 生成了 %d 个资源的 YAML", name, len(parts))
	return result, nil
}

// --- 解析辅助函数 ---

func parseStringList(data json.RawMessage) []string {
	if len(data) == 0 {
		return nil
	}
	var list []string
	_ = json.Unmarshal(data, &list)
	return list
}

type ProbeCfg struct {
	Type             string `json:"type"`
	Path             string `json:"path,omitempty"`
	Port             int    `json:"port,omitempty"`
	Command          string `json:"command,omitempty"`
	InitialDelaySecs int    `json:"initialDelaySeconds,omitempty"`
	PeriodSecs       int    `json:"periodSeconds,omitempty"`
}

func parseProbe(data json.RawMessage) *corev1.Probe {
	if len(data) == 0 || string(data) == "{}" || string(data) == "null" {
		return nil
	}
	var cfg ProbeCfg
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Type == "" {
		return nil
	}
	probe := &corev1.Probe{
		InitialDelaySeconds: int32(cfg.InitialDelaySecs),
		PeriodSeconds:       int32(cfg.PeriodSecs),
	}
	if probe.PeriodSeconds == 0 {
		probe.PeriodSeconds = 10
	}
	switch cfg.Type {
	case "http":
		probe.ProbeHandler = corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{Path: cfg.Path, Port: intstr.FromInt(cfg.Port)},
		}
	case "tcp":
		probe.ProbeHandler = corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(cfg.Port)},
		}
	case "exec":
		probe.ProbeHandler = corev1.ProbeHandler{
			Exec: &corev1.ExecAction{Command: strings.Split(cfg.Command, " ")},
		}
	}
	return probe
}

func buildResources(cpuReq, memReq, cpuLim, memLim string) corev1.ResourceRequirements {
	res := corev1.ResourceRequirements{}
	if cpuReq != "" || memReq != "" {
		res.Requests = corev1.ResourceList{}
		if cpuReq != "" {
			res.Requests[corev1.ResourceCPU] = resource.MustParse(cpuReq)
		}
		if memReq != "" {
			res.Requests[corev1.ResourceMemory] = resource.MustParse(memReq)
		}
	}
	if cpuLim != "" || memLim != "" {
		res.Limits = corev1.ResourceList{}
		if cpuLim != "" {
			res.Limits[corev1.ResourceCPU] = resource.MustParse(cpuLim)
		}
		if memLim != "" {
			res.Limits[corev1.ResourceMemory] = resource.MustParse(memLim)
		}
	}
	return res
}

// resolveImageRepo 从 deployment 或 service 获取镜像仓库路径
func resolveImageRepo(dep *model.Deployment, svc *model.Service) string {
	if dep.ExternalImage != "" {
		return ""
	}
	// 优先从 Build 记录读取镜像路径
	if dep.Build != nil && dep.Build.ImageRepo != "" {
		return dep.Build.ImageRepo
	}
	if svc != nil && svc.ImageRepo != "" {
		return svc.ImageRepo
	}
	return ""
}

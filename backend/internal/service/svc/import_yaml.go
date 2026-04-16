package svc

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ImportedService 解析后的服务信息
type ImportedService struct {
	Name       string   `json:"name" yaml:"name"`
	Namespace  string   `json:"namespace" yaml:"namespace"`
	Image      string   `json:"image" yaml:"image"`
	Replicas   int      `json:"replicas" yaml:"replicas"`
	Port       int      `json:"port" yaml:"port"`
	CPURequest string   `json:"cpu_request,omitempty"`
	MemRequest string   `json:"mem_request,omitempty"`
	CPULimit   string   `json:"cpu_limit,omitempty"`
	MemLimit   string   `json:"mem_limit,omitempty"`
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors,omitempty"`
}

// k8sDeployment 用于解析 K8s YAML
type k8sDeployment struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Replicas int `yaml:"replicas"`
		Template struct {
			Spec struct {
				Containers []struct {
					Name  string `yaml:"name"`
					Image string `yaml:"image"`
					Ports []struct {
						ContainerPort int `yaml:"containerPort"`
					} `yaml:"ports"`
					Resources struct {
						Requests struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"requests"`
						Limits struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"limits"`
					} `yaml:"resources"`
				} `yaml:"containers"`
			} `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

// ParseK8sYAML 解析 K8s Deployment YAML
func ParseK8sYAML(reader io.Reader) ([]ImportedService, error) {
	decoder := yaml.NewDecoder(reader)
	var results []ImportedService

	for {
		var dep k8sDeployment
		err := decoder.Decode(&dep)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("YAML 解析失败: %w", err)
		}

		if dep.Kind != "Deployment" && dep.Kind != "" {
			continue
		}

		svc := ImportedService{
			Name:      dep.Metadata.Name,
			Namespace: dep.Metadata.Namespace,
			Replicas:  dep.Spec.Replicas,
			Valid:     true,
		}

		if svc.Namespace == "" {
			svc.Namespace = "default"
		}
		if svc.Replicas == 0 {
			svc.Replicas = 1
		}

		if len(dep.Spec.Template.Spec.Containers) > 0 {
			c := dep.Spec.Template.Spec.Containers[0]
			svc.Image = c.Image
			if len(c.Ports) > 0 {
				svc.Port = c.Ports[0].ContainerPort
			}
			svc.CPURequest = c.Resources.Requests.CPU
			svc.MemRequest = c.Resources.Requests.Memory
			svc.CPULimit = c.Resources.Limits.CPU
			svc.MemLimit = c.Resources.Limits.Memory
		}

		if svc.Name == "" {
			svc.Valid = false
			svc.Errors = append(svc.Errors, "缺少服务名称")
		}
		if svc.Image == "" {
			svc.Valid = false
			svc.Errors = append(svc.Errors, "缺少容器镜像")
		}

		results = append(results, svc)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到有效的 Deployment 定义")
	}
	return results, nil
}

// ParseImportFile 根据内容类型解析导入文件
func ParseImportFile(reader io.Reader, contentType string) ([]ImportedService, error) {
	return ParseK8sYAML(reader)
}

package deploy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnvImageInfo 从 app-env.yaml 解析出的镜像信息
type EnvImageInfo struct {
	Repository      string `json:"repository"`
	Tag             string `json:"tag"`
	ImagePullPolicy string `json:"image_pull_policy"`
	FullImage       string `json:"full_image"`
}

// helmValuesWithImage 解析 image 段的 YAML 结构
type helmValuesWithImage struct {
	Image struct {
		Repository      string      `yaml:"repository"`
		Tag             interface{} `yaml:"tag"` // 可能是 string 或 number
		ImagePullPolicy string      `yaml:"imagePullPolicy"`
	} `yaml:"image"`
}

// ParseEnvImage 解析 Helm values YAML 提取 image 字段
func ParseEnvImage(content string) (*EnvImageInfo, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("YAML 内容为空")
	}

	var vals helmValuesWithImage
	if err := yaml.Unmarshal([]byte(content), &vals); err != nil {
		return nil, fmt.Errorf("YAML 解析失败: %w", err)
	}

	if vals.Image.Repository == "" {
		return nil, fmt.Errorf("YAML 中未找到 image.repository 字段")
	}

	tag := fmt.Sprintf("%v", vals.Image.Tag)
	if tag == "" || tag == "<nil>" {
		tag = "latest"
	}

	return &EnvImageInfo{
		Repository:      vals.Image.Repository,
		Tag:             tag,
		ImagePullPolicy: vals.Image.ImagePullPolicy,
		FullImage:       fmt.Sprintf("%s:%s", vals.Image.Repository, tag),
	}, nil
}

package pkg

import "strings"

// ParseImageRef 解析完整镜像地址为 repository 和 tag
// docker.io/myorg/myapp:v1.2.3 → ("docker.io/myorg/myapp", "v1.2.3")
// nginx → ("nginx", "latest")
// registry.io/app@sha256:abc → ("registry.io/app", "sha256:abc")
func ParseImageRef(ref string) (repository, tag string) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", ""
	}

	// digest 格式: repo@sha256:xxx
	if idx := strings.Index(ref, "@"); idx > 0 {
		return ref[:idx], ref[idx+1:]
	}

	// 找最后一个冒号，但要排除 registry 端口号中的冒号
	// 策略：如果冒号后面没有 /，则认为是 tag 分隔符
	lastColon := strings.LastIndex(ref, ":")
	if lastColon < 0 {
		return ref, "latest"
	}

	afterColon := ref[lastColon+1:]
	if strings.Contains(afterColon, "/") {
		// 冒号后面有 /，说明是端口号（如 registry.io:5000/app）
		return ref, "latest"
	}

	return ref[:lastColon], afterColon
}

// ValidateImageRef 校验镜像地址基本格式
func ValidateImageRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	return ref != ""
}

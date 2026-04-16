package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRepo string
		wantTag  string
	}{
		{"完整地址", "docker.io/myorg/myapp:v1.2.3", "docker.io/myorg/myapp", "v1.2.3"},
		{"简单地址", "myapp:latest", "myapp", "latest"},
		{"无 tag", "nginx", "nginx", "latest"},
		{"带端口的 registry", "registry.io:5000/myapp:v1", "registry.io:5000/myapp", "v1"},
		{"digest", "registry.io/app@sha256:abc123", "registry.io/app", "sha256:abc123"},
		{"空字符串", "", "", ""},
		{"私有仓库", "123456.dkr.ecr.us-east-1.amazonaws.com/myapp:build-42", "123456.dkr.ecr.us-east-1.amazonaws.com/myapp", "build-42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, tag := ParseImageRef(tt.input)
			assert.Equal(t, tt.wantRepo, repo)
			assert.Equal(t, tt.wantTag, tag)
		})
	}
}

func TestValidateImageRef(t *testing.T) {
	assert.True(t, ValidateImageRef("docker.io/myorg/myapp:v1.2.3"))
	assert.True(t, ValidateImageRef("nginx:latest"))
	assert.True(t, ValidateImageRef("myapp"))
	assert.False(t, ValidateImageRef(""))
	assert.False(t, ValidateImageRef("   "))
}

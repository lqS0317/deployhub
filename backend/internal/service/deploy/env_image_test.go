package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnvImage_Standard(t *testing.T) {
	yaml := `
image:
  repository: docker.io/myorg/myapp
  tag: v1.2.3
  imagePullPolicy: IfNotPresent
`
	info, err := ParseEnvImage(yaml)
	require.NoError(t, err)
	assert.Equal(t, "docker.io/myorg/myapp", info.Repository)
	assert.Equal(t, "v1.2.3", info.Tag)
	assert.Equal(t, "IfNotPresent", info.ImagePullPolicy)
	assert.Equal(t, "docker.io/myorg/myapp:v1.2.3", info.FullImage)
}

func TestParseEnvImage_MissingTag(t *testing.T) {
	yaml := `
image:
  repository: nginx
`
	info, err := ParseEnvImage(yaml)
	require.NoError(t, err)
	assert.Equal(t, "nginx", info.Repository)
	assert.Equal(t, "latest", info.Tag)
	assert.Equal(t, "nginx:latest", info.FullImage)
}

func TestParseEnvImage_NoImageSection(t *testing.T) {
	yaml := `
replicaCount: 3
service:
  port: 8080
`
	_, err := ParseEnvImage(yaml)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image")
}

func TestParseEnvImage_EmptyContent(t *testing.T) {
	_, err := ParseEnvImage("")
	assert.Error(t, err)
}

func TestParseEnvImage_NumericTag(t *testing.T) {
	yaml := `
image:
  repository: blockscout/blockscout-optimism
  tag: 7.0.2
  imagePullPolicy: IfNotPresent
`
	info, err := ParseEnvImage(yaml)
	require.NoError(t, err)
	assert.Equal(t, "blockscout/blockscout-optimism", info.Repository)
	assert.Equal(t, "7.0.2", info.Tag)
	assert.Equal(t, "blockscout/blockscout-optimism:7.0.2", info.FullImage)
}

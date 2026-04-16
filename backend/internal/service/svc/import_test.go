package svc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseK8sDeploymentYAML(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: production
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: my-app
        image: registry.example.com/my-app:v1.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi`

	results, err := ParseK8sYAML(strings.NewReader(yaml))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "my-app", results[0].Name)
	assert.Equal(t, "production", results[0].Namespace)
	assert.Equal(t, 3, results[0].Replicas)
	assert.Equal(t, 8080, results[0].Port)
	assert.Equal(t, "registry.example.com/my-app:v1.0", results[0].Image)
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := ParseK8sYAML(strings.NewReader("not: valid: yaml: [[["))
	assert.Error(t, err)
}

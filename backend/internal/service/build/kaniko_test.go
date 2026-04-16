package build

import (
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKanikoJob(t *testing.T) {
	build := &model.Build{
		ID:        42,
		ServiceID: 1,
		GitBranch: "main",
		GitCommit: "abc123def",
		ImageTag:  "registry.example.com/test-svc:main-1700000000",
	}
	service := &model.Service{
		ID:             1,
		DockerfilePath: "./Dockerfile",
		GitRepo: &model.GitRepo{
			URL: "https://github.com/example/repo.git",
		},
	}

	job := GenerateKanikoJob(build, service, "git-token-secret", "registry-auth-base64", "deployhub-builds")

	t.Run("Job 基本属性", func(t *testing.T) {
		require.NotNil(t, job)
		assert.Equal(t, "kaniko-build-42", job.Name)
		assert.Equal(t, "deployhub-builds", job.Namespace)
		assert.NotNil(t, job.Spec.BackoffLimit)
		assert.Equal(t, int32(0), *job.Spec.BackoffLimit)
	})

	t.Run("Pod 模板配置", func(t *testing.T) {
		podSpec := job.Spec.Template.Spec
		assert.Equal(t, "Never", string(podSpec.RestartPolicy))
		assert.Len(t, podSpec.InitContainers, 1)
		assert.Len(t, podSpec.Containers, 1)
	})

	t.Run("Init 容器使用 alpine/git 拉取代码", func(t *testing.T) {
		initContainer := job.Spec.Template.Spec.InitContainers[0]
		assert.Equal(t, "git-clone", initContainer.Name)
		assert.Equal(t, "alpine/git:latest", initContainer.Image)
	})

	t.Run("主容器使用 Kaniko executor", func(t *testing.T) {
		mainContainer := job.Spec.Template.Spec.Containers[0]
		assert.Equal(t, "kaniko", mainContainer.Name)
		assert.Equal(t, "gcr.io/kaniko-project/executor:latest", mainContainer.Image)

		argsStr := ""
		for _, a := range mainContainer.Args {
			argsStr += a + " "
		}
		assert.Contains(t, argsStr, "--dockerfile=./Dockerfile")
		assert.Contains(t, argsStr, "--context=dir:///workspace")
		assert.Contains(t, argsStr, "--destination=registry.example.com/test-svc:main-1700000000")
	})

	t.Run("挂载了 registry auth 卷", func(t *testing.T) {
		podSpec := job.Spec.Template.Spec
		hasDockerConfigVolume := false
		for _, v := range podSpec.Volumes {
			if v.Name == "docker-config" {
				hasDockerConfigVolume = true
			}
		}
		assert.True(t, hasDockerConfigVolume)
	})

	t.Run("设置了资源限制", func(t *testing.T) {
		mainContainer := job.Spec.Template.Spec.Containers[0]
		assert.NotNil(t, mainContainer.Resources.Limits)
		assert.NotNil(t, mainContainer.Resources.Requests)
	})
}

func TestGenerateKanikoJob_CustomNamespace(t *testing.T) {
	build := &model.Build{ID: 1}
	service := &model.Service{
		GitRepo: &model.GitRepo{URL: "https://github.com/example/repo.git"},
	}

	job := GenerateKanikoJob(build, service, "cred", "auth", "custom-ns")
	assert.Equal(t, "custom-ns", job.Namespace)
}

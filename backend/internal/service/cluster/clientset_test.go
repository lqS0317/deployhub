package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildClientsetInvalidKubeconfig(t *testing.T) {
	_, err := BuildClientset("invalid-kubeconfig")
	assert.Error(t, err)
}

func TestBuildClientsetEmptyKubeconfig(t *testing.T) {
	_, err := BuildClientset("")
	assert.Error(t, err)
}

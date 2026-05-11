package routing

import (
	"errors"
	"fmt"

	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// ErrNamespaceNotMapped namespace 未在集群映射表中登记时返回的错误。
// 上层 handler 可通过 errors.Is 判断并转 400。
var ErrNamespaceNotMapped = errors.New("namespace 未在集群映射列表中")

// validateClusterNamespace 校验给定 namespace 是否属于该集群的映射列表。
// 若未注入 ClusterNamespaceRepository（如测试场景），跳过校验保持向后兼容。
func validateClusterNamespace(repo repository.ClusterNamespaceRepository, clusterID uint, namespace string) error {
	if repo == nil {
		return nil
	}
	if namespace == "" {
		return fmt.Errorf("%w：命名空间不能为空，请先在集群管理中配置 namespace 映射", ErrNamespaceNotMapped)
	}
	_, err := repo.FindByClusterAndNamespace(clusterID, namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w：%q 未在集群映射列表中，请先在集群管理中配置", ErrNamespaceNotMapped, namespace)
		}
		return fmt.Errorf("校验命名空间映射失败: %w", err)
	}
	return nil
}

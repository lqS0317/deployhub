package notification

import (
	"errors"
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotificationRepo 通知仓储 mock 实现
type mockNotificationRepo struct {
	notifications map[uint]*model.Notification
	nextID        uint
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{notifications: make(map[uint]*model.Notification), nextID: 1}
}

func (m *mockNotificationRepo) Create(n *model.Notification) error {
	n.ID = m.nextID
	m.nextID++
	m.notifications[n.ID] = n
	return nil
}

func (m *mockNotificationRepo) FindByID(id uint) (*model.Notification, error) {
	if n, ok := m.notifications[id]; ok {
		return n, nil
	}
	return nil, errors.New("记录不存在")
}

func (m *mockNotificationRepo) Update(n *model.Notification) error {
	if _, ok := m.notifications[n.ID]; !ok {
		return errors.New("记录不存在")
	}
	m.notifications[n.ID] = n
	return nil
}

func (m *mockNotificationRepo) List(userID uint, page, pageSize int, isRead *bool) ([]model.Notification, int64, error) {
	var result []model.Notification
	for _, n := range m.notifications {
		if n.UserID != userID {
			continue
		}
		if isRead != nil && n.IsRead != *isRead {
			continue
		}
		result = append(result, *n)
	}
	total := int64(len(result))
	start := (page - 1) * pageSize
	if start >= len(result) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (m *mockNotificationRepo) MarkAllRead(userID uint) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationRepo) UnreadCount(userID uint) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

var _ repository.NotificationRepository = (*mockNotificationRepo)(nil)

func TestCreate(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	t.Run("成功创建通知", func(t *testing.T) {
		n, err := svc.Create(1, "build_complete", "构建完成", "服务 test-svc 构建成功", "build", 10)
		require.NoError(t, err)
		assert.Equal(t, uint(1), n.UserID)
		assert.Equal(t, "build_complete", n.Type)
		assert.Equal(t, "构建完成", n.Title)
		assert.Equal(t, "服务 test-svc 构建成功", n.Content)
		assert.Equal(t, "build", n.ReferenceType)
		assert.Equal(t, uint(10), n.ReferenceID)
		assert.False(t, n.IsRead)
	})
}

func TestListWithReadFilter(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	_, _ = svc.Create(1, "build_complete", "构建完成1", "", "", 0)
	n2, _ := svc.Create(1, "deploy_result", "部署结果", "", "", 0)
	_, _ = svc.Create(2, "build_complete", "其他用户通知", "", "", 0)

	// 标记一条为已读
	_ = svc.MarkRead(n2.ID, 1)

	t.Run("列出用户全部通知", func(t *testing.T) {
		items, total, err := svc.List(1, 1, 20, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, items, 2)
	})

	t.Run("仅列出未读通知", func(t *testing.T) {
		isRead := false
		items, total, err := svc.List(1, 1, 20, &isRead)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
		assert.False(t, items[0].IsRead)
	})

	t.Run("仅列出已读通知", func(t *testing.T) {
		isRead := true
		items, total, err := svc.List(1, 1, 20, &isRead)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
		assert.True(t, items[0].IsRead)
	})
}

func TestMarkRead(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	n, _ := svc.Create(1, "build_complete", "构建完成", "", "", 0)

	t.Run("标记自己的通知为已读", func(t *testing.T) {
		err := svc.MarkRead(n.ID, 1)
		require.NoError(t, err)
		assert.True(t, repo.notifications[n.ID].IsRead)
	})

	t.Run("标记不属于自己的通知应失败", func(t *testing.T) {
		err := svc.MarkRead(n.ID, 999)
		assert.Error(t, err)
	})

	t.Run("标记不存在的通知应失败", func(t *testing.T) {
		err := svc.MarkRead(999, 1)
		assert.Error(t, err)
	})
}

func TestMarkAllRead(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	_, _ = svc.Create(1, "build_complete", "通知1", "", "", 0)
	_, _ = svc.Create(1, "deploy_result", "通知2", "", "", 0)
	_, _ = svc.Create(2, "build_complete", "其他用户通知", "", "", 0)

	t.Run("标记全部已读", func(t *testing.T) {
		count, err := svc.MarkAllRead(1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("再次标记全部已读应返回 0", func(t *testing.T) {
		count, err := svc.MarkAllRead(1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestUnreadCount(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	_, _ = svc.Create(1, "build_complete", "通知1", "", "", 0)
	_, _ = svc.Create(1, "deploy_result", "通知2", "", "", 0)
	n3, _ := svc.Create(1, "approval_request", "通知3", "", "", 0)

	t.Run("未读数量为 3", func(t *testing.T) {
		count, err := svc.UnreadCount(1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("标记一条已读后未读数量为 2", func(t *testing.T) {
		_ = svc.MarkRead(n3.ID, 1)
		count, err := svc.UnreadCount(1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

func TestNotifyBuildComplete(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	err := svc.NotifyBuildComplete(1, 10, "test-svc", "success")
	require.NoError(t, err)

	assert.Len(t, repo.notifications, 1)
	n := repo.notifications[1]
	assert.Equal(t, "build_complete", n.Type)
	assert.Equal(t, "build", n.ReferenceType)
	assert.Equal(t, uint(10), n.ReferenceID)
}

func TestNotifyDeployResult(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	err := svc.NotifyDeployResult(1, 20, "test-svc", "success")
	require.NoError(t, err)

	n := repo.notifications[1]
	assert.Equal(t, "deploy_result", n.Type)
	assert.Equal(t, "deployment", n.ReferenceType)
	assert.Equal(t, uint(20), n.ReferenceID)
}

func TestNotifyApprovalRequest(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := NewNotificationService(repo)

	err := svc.NotifyApprovalRequest(5, 30, "test-svc", "张三")
	require.NoError(t, err)

	n := repo.notifications[1]
	assert.Equal(t, uint(5), n.UserID)
	assert.Equal(t, "approval_request", n.Type)
	assert.Equal(t, "deployment", n.ReferenceType)
	assert.Equal(t, uint(30), n.ReferenceID)
}

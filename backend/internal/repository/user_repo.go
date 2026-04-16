package repository

import "deployhub/internal/model"

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByOAuth(provider, oauthID string) (*model.User, error)
	Update(user *model.User) error
	List(page, pageSize int) ([]model.User, int64, error)
	UpdateRole(id uint, role string) error
	UpdateStatus(id uint, status string) error
	FindByRole(role string) ([]model.User, error)
}

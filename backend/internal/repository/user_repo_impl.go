package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByOAuth(provider, oauthID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *userRepository) UpdateRole(id uint, role string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *userRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) FindByRole(role string) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("role = ? AND status = 'active'", role).Find(&users).Error
	return users, err
}

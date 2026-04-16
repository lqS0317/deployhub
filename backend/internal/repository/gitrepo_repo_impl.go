package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type gitRepoRepository struct {
	db *gorm.DB
}

// NewGitRepoRepository 创建 Git 仓库数据访问实例
func NewGitRepoRepository(db *gorm.DB) GitRepoRepository {
	return &gitRepoRepository{db: db}
}

func (r *gitRepoRepository) Create(repo *model.GitRepo) error {
	return r.db.Create(repo).Error
}

func (r *gitRepoRepository) FindByID(id uint) (*model.GitRepo, error) {
	var repo model.GitRepo
	err := r.db.First(&repo, id).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *gitRepoRepository) FindByName(name string) (*model.GitRepo, error) {
	var repo model.GitRepo
	err := r.db.Where("name = ?", name).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *gitRepoRepository) Update(repo *model.GitRepo) error {
	return r.db.Save(repo).Error
}

func (r *gitRepoRepository) Delete(id uint) error {
	return r.db.Delete(&model.GitRepo{}, id).Error
}

func (r *gitRepoRepository) List(page, pageSize int) ([]model.GitRepo, int64, error) {
	var repos []model.GitRepo
	var total int64
	r.db.Model(&model.GitRepo{}).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&repos).Error
	return repos, total, err
}

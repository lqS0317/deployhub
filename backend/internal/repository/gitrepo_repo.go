package repository

import "deployhub/internal/model"

// GitRepoRepository Git 仓库数据访问接口
type GitRepoRepository interface {
	Create(repo *model.GitRepo) error
	FindByID(id uint) (*model.GitRepo, error)
	FindByName(name string) (*model.GitRepo, error)
	Update(repo *model.GitRepo) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.GitRepo, int64, error)
}

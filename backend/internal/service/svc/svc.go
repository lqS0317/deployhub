package svc

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// ServiceService 服务管理
type ServiceService struct {
	svcRepo    repository.ServiceRepository
	memberRepo repository.ServiceMemberRepository
}

func NewServiceService(svcRepo repository.ServiceRepository, memberRepo repository.ServiceMemberRepository) *ServiceService {
	return &ServiceService{svcRepo: svcRepo, memberRepo: memberRepo}
}

// Create 创建服务并自动添加 owner 成员
func (s *ServiceService) Create(svc *model.Service) (*model.Service, error) {
	if _, err := s.svcRepo.FindByName(svc.Name); err == nil {
		return nil, fmt.Errorf("服务 %s 已存在", svc.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询服务失败: %w", err)
	}
	if err := s.svcRepo.Create(svc); err != nil {
		return nil, fmt.Errorf("创建服务失败: %w", err)
	}
	member := &model.ServiceMember{ServiceID: svc.ID, UserID: svc.OwnerID, Role: "owner"}
	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("创建服务 owner 失败: %w", err)
	}
	return svc, nil
}

func (s *ServiceService) GetByID(id uint) (*model.Service, error) { return s.svcRepo.FindByID(id) }

func (s *ServiceService) List(page, pageSize int) ([]model.Service, int64, error) {
	return s.svcRepo.List(page, pageSize)
}

func (s *ServiceService) Update(svc *model.Service) (*model.Service, error) {
	if err := s.svcRepo.Update(svc); err != nil {
		return nil, fmt.Errorf("更新服务失败: %w", err)
	}
	// 重新查询以带上最新的关联信息（如 GitRepo），避免返回旧的预加载数据
	updated, err := s.svcRepo.FindByID(svc.ID)
	if err != nil {
		return nil, fmt.Errorf("查询更新后的服务失败: %w", err)
	}
	return updated, nil
}

func (s *ServiceService) Delete(id uint) error { return s.svcRepo.Delete(id) }

func (s *ServiceService) BatchCreate(services []*model.Service) error {
	return s.svcRepo.BatchCreate(services)
}

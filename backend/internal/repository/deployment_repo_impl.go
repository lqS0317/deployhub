package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type deploymentRepository struct {
	db *gorm.DB
}

// NewDeploymentRepository 创建部署记录仓储实例
func NewDeploymentRepository(db *gorm.DB) DeploymentRepository {
	return &deploymentRepository{db: db}
}

func (r *deploymentRepository) Create(deployment *model.Deployment) error {
	return r.db.Create(deployment).Error
}

func (r *deploymentRepository) FindByID(id uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Preload("Service").Preload("TriggerUser").Preload("Build").Preload("Cluster").First(&deployment, id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *deploymentRepository) Update(deployment *model.Deployment) error {
	return r.db.Save(deployment).Error
}

func (r *deploymentRepository) List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error) {
	var deployments []model.Deployment
	var total int64

	query := r.db.Model(&model.Deployment{})
	if serviceID != nil {
		query = query.Where("service_id = ?", *serviceID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Service").Preload("TriggerUser").Preload("Cluster").
		Offset(offset).Limit(pageSize).Order("id DESC").Find(&deployments).Error
	return deployments, total, err
}

// FindActiveByService 查找指定服务正在进行中的部署（pending_approval / approved / deploying）
func (r *deploymentRepository) FindActiveByService(serviceID uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Where("service_id = ? AND status IN ?", serviceID, []string{
		model.DeployStatusPendingApproval,
		model.DeployStatusApproved,
		model.DeployStatusPreviewing,
		model.DeployStatusPreviewed,
		model.DeployStatusDeploying,
		model.DeployStatusPodChecking,
	}).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// FindLastSuccessful 查找指定服务最近一次成功的部署
func (r *deploymentRepository) FindLastSuccessful(serviceID uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Where("service_id = ? AND status IN ?", serviceID, []string{
		model.DeployStatusSuccess,
		model.DeployStatusPodHealthy,
	}).Order("id DESC").First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *deploymentRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Deployment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *deploymentRepository) UpdateStatusWithReason(id uint, status, reason string) error {
	return r.db.Model(&model.Deployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"fail_reason": reason,
	}).Error
}

func (r *deploymentRepository) UpdatePodStatus(id uint, status, podStatus, podMessage string) error {
	return r.db.Model(&model.Deployment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"pod_status":  podStatus,
		"pod_message": podMessage,
	}).Error
}

func (r *deploymentRepository) UpdateField(id uint, field string, value interface{}) error {
	return r.db.Model(&model.Deployment{}).Where("id = ?", id).Update(field, value).Error
}

func (r *deploymentRepository) FindByStatuses(statuses []string) ([]model.Deployment, error) {
	var deployments []model.Deployment
	err := r.db.Preload("Service").Where("status IN ?", statuses).Find(&deployments).Error
	return deployments, err
}

func (r *deploymentRepository) Delete(id uint) error {
	// 级联删除关联审批记录
	r.db.Where("deployment_id = ?", id).Delete(&model.Approval{})
	return r.db.Delete(&model.Deployment{}, id).Error
}

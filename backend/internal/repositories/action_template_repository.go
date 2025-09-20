package repositories

import (
	"gorm.io/gorm"
	"smart-device-management/internal/models"
)

// ActionTemplateRepository 动作模板仓库接口
type ActionTemplateRepository interface {
	Create(template *models.ActionTemplate) error
	GetByID(id uint) (*models.ActionTemplate, error)
	GetAll() ([]models.ActionTemplate, error)
	GetByType(templateType string) ([]models.ActionTemplate, error)
	Update(template *models.ActionTemplate) error
	Delete(id uint) error
}

// actionTemplateRepository 动作模板仓库实现
type actionTemplateRepository struct {
	db *gorm.DB
}

// NewActionTemplateRepository 创建动作模板仓库
func NewActionTemplateRepository(db *gorm.DB) ActionTemplateRepository {
	return &actionTemplateRepository{db: db}
}

// Create 创建动作模板
func (r *actionTemplateRepository) Create(template *models.ActionTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取动作模板
func (r *actionTemplateRepository) GetByID(id uint) (*models.ActionTemplate, error) {
	var template models.ActionTemplate
	err := r.db.First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetAll 获取所有动作模板
func (r *actionTemplateRepository) GetAll() ([]models.ActionTemplate, error) {
	var templates []models.ActionTemplate
	err := r.db.Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// GetByType 根据类型获取动作模板
func (r *actionTemplateRepository) GetByType(templateType string) ([]models.ActionTemplate, error) {
	var templates []models.ActionTemplate
	err := r.db.Where("type = ?", templateType).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// Update 更新动作模板
func (r *actionTemplateRepository) Update(template *models.ActionTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除动作模板（软删除）
func (r *actionTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.ActionTemplate{}, id).Error
}

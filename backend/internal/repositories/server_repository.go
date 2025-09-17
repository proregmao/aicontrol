package repositories

import (
	"smart-device-management/internal/models"
	"time"

	"gorm.io/gorm"
)

// ServerRepository 服务器仓储接口
type ServerRepository interface {
	FindAll() ([]models.Server, error)
	FindByID(id uint) (*models.Server, error)
	FindByIPAddress(ipAddress string) (*models.Server, error)
	FindByStatus(status models.ServerStatus) ([]models.Server, error)
	Create(server *models.Server) error
	Update(server *models.Server) error
	UpdateStatus(id uint, status models.ServerStatus, connected bool) error
	Delete(id uint) error
	Count() (int64, error)
}

// serverRepository 服务器仓储实现
type serverRepository struct {
	db *gorm.DB
}

// NewServerRepository 创建服务器仓储实例
func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepository{
		db: db,
	}
}

// FindAll 获取所有服务器
func (r *serverRepository) FindAll() ([]models.Server, error) {
	var servers []models.Server
	// 移除 Preload("Device") 避免关联查询问题
	err := r.db.Find(&servers).Error
	return servers, err
}

// FindByID 根据ID获取服务器
func (r *serverRepository) FindByID(id uint) (*models.Server, error) {
	var server models.Server
	// 移除 Preload("Device") 避免关联查询问题
	err := r.db.Where("id = ?", id).First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// FindByIPAddress 根据IP地址获取服务器
func (r *serverRepository) FindByIPAddress(ipAddress string) (*models.Server, error) {
	var server models.Server
	err := r.db.Where("ip_address = ?", ipAddress).First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

// FindByStatus 根据状态获取服务器列表
func (r *serverRepository) FindByStatus(status models.ServerStatus) ([]models.Server, error) {
	var servers []models.Server
	err := r.db.Where("status = ?", status).Find(&servers).Error
	return servers, err
}

// Create 创建服务器
func (r *serverRepository) Create(server *models.Server) error {
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	return r.db.Create(server).Error
}

// Update 更新服务器
func (r *serverRepository) Update(server *models.Server) error {
	server.UpdatedAt = time.Now()
	return r.db.Save(server).Error
}

// UpdateStatus 更新服务器状态
func (r *serverRepository) UpdateStatus(id uint, status models.ServerStatus, connected bool) error {
	return r.db.Model(&models.Server{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"connected":  connected,
			"updated_at": time.Now(),
		}).Error
}

// Delete 删除服务器（软删除）
func (r *serverRepository) Delete(id uint) error {
	return r.db.Delete(&models.Server{}, id).Error
}

// Count 获取服务器总数
func (r *serverRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Server{}).Count(&count).Error
	return count, err
}

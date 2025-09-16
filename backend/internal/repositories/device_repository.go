package repositories

import (
	"smart-device-management/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{
		db: db,
	}
}

// FindAll 获取所有设备
func (r *DeviceRepository) FindAll() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Find(&devices).Error
	return devices, err
}

// FindByID 根据ID获取设备
func (r *DeviceRepository) FindByID(id uint) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("id = ?", id).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// FindByName 根据名称获取设备
func (r *DeviceRepository) FindByName(name string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("device_name = ?", name).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// FindByType 根据类型获取设备
func (r *DeviceRepository) FindByType(deviceType string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("device_type = ?", deviceType).Find(&devices).Error
	return devices, err
}

// FindByStatus 根据状态获取设备
func (r *DeviceRepository) FindByStatus(status string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("status = ?", status).Find(&devices).Error
	return devices, err
}

// Create 创建设备
func (r *DeviceRepository) Create(device *models.Device) error {
	return r.db.Create(device).Error
}

// Update 更新设备
func (r *DeviceRepository) Update(device *models.Device) error {
	device.UpdatedAt = time.Now()
	return r.db.Save(device).Error
}

// UpdateStatus 更新设备状态
func (r *DeviceRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Device{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// Delete 删除设备
func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Device{}, id).Error
}

// GetDeviceCount 获取设备总数
func (r *DeviceRepository) GetDeviceCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Device{}).Count(&count).Error
	return count, err
}

// GetDeviceCountByType 根据类型获取设备数量
func (r *DeviceRepository) GetDeviceCountByType(deviceType string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Device{}).Where("device_type = ?", deviceType).Count(&count).Error
	return count, err
}

// GetDeviceCountByStatus 根据状态获取设备数量
func (r *DeviceRepository) GetDeviceCountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Device{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// FindWithPagination 分页获取设备
func (r *DeviceRepository) FindWithPagination(offset, limit int) ([]models.Device, int64, error) {
	var devices []models.Device
	var total int64

	// 获取总数
	err := r.db.Model(&models.Device{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = r.db.Offset(offset).Limit(limit).Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// FindByTypeWithPagination 根据类型分页获取设备
func (r *DeviceRepository) FindByTypeWithPagination(deviceType string, offset, limit int) ([]models.Device, int64, error) {
	var devices []models.Device
	var total int64

	query := r.db.Model(&models.Device{}).Where("device_type = ?", deviceType)

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Offset(offset).Limit(limit).Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// SearchDevices 搜索设备
func (r *DeviceRepository) SearchDevices(keyword string) ([]models.Device, error) {
	var devices []models.Device
	searchPattern := "%" + keyword + "%"

	err := r.db.Where("device_name ILIKE ? OR location ILIKE ?", searchPattern, searchPattern).
		Find(&devices).Error

	return devices, err
}

// GetActiveDevices 获取活跃设备
func (r *DeviceRepository) GetActiveDevices() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("status = ?", "active").Find(&devices).Error
	return devices, err
}

// GetDevicesByLocation 根据位置获取设备
func (r *DeviceRepository) GetDevicesByLocation(location string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("location ILIKE ?", "%"+location+"%").Find(&devices).Error
	return devices, err
}

// BatchUpdateStatus 批量更新设备状态
func (r *DeviceRepository) BatchUpdateStatus(ids []uuid.UUID, status string) error {
	return r.db.Model(&models.Device{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// GetDeviceStatistics 获取设备统计信息
func (r *DeviceRepository) GetDeviceStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总设备数
	var totalCount int64
	err := r.db.Model(&models.Device{}).Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	stats["total"] = totalCount

	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	err = r.db.Model(&models.Device{}).
		Select("device_type as type, COUNT(*) as count").
		Group("device_type").
		Scan(&typeStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_type"] = typeStats

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	err = r.db.Model(&models.Device{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_status"] = statusStats

	return stats, nil
}

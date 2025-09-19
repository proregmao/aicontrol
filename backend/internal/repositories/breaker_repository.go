package repositories

import (
	"smart-device-management/internal/models"

	"gorm.io/gorm"
)

// BreakerRepository 断路器仓库接口
type BreakerRepository interface {
	GetAll() ([]models.Breaker, error)
	GetByID(id uint) (*models.Breaker, error)
	GetByIPAddress(ipAddress string, port int) (*models.Breaker, error)
	GetEnabledBreakers() ([]*models.Breaker, error)
	Create(breaker *models.Breaker, device *models.Device) error
	Update(breaker *models.Breaker) error
	Delete(id uint) error
	UpdateBreakerLockStatus(id uint, isLocked bool) error

	// 控制相关
	CreateControl(control *models.BreakerControl) error
	UpdateControl(control *models.BreakerControl) error
	GetControl(controlID string) (*models.BreakerControl, error)

	// 绑定相关
	GetBindings(breakerID uint) ([]models.BreakerServerBinding, error)
	CreateBinding(binding *models.BreakerServerBinding) error
	UpdateBinding(binding *models.BreakerServerBinding) error
	DeleteBinding(id uint) error
	GetBinding(id uint) (*models.BreakerServerBinding, error)
}

// breakerRepository 断路器仓库实现
type breakerRepository struct {
	db *gorm.DB
}

// NewBreakerRepository 创建断路器仓库
func NewBreakerRepository(db *gorm.DB) BreakerRepository {
	return &breakerRepository{db: db}
}

// GetAll 获取所有断路器
func (r *breakerRepository) GetAll() ([]models.Breaker, error) {
	var breakers []models.Breaker
	err := r.db.Preload("Device").
		Preload("Bindings").
		Preload("Bindings.Server").
		Order("id ASC").  // 按ID升序排序，确保添加先后顺序
		Find(&breakers).Error
	return breakers, err
}

// GetByID 根据ID获取断路器
func (r *breakerRepository) GetByID(id uint) (*models.Breaker, error) {
	var breaker models.Breaker
	err := r.db.Preload("Device").
		Preload("Bindings").
		Preload("Bindings.Server").
		First(&breaker, id).Error
	return &breaker, err
}

// GetByIPAddress 根据IP地址获取断路器
func (r *breakerRepository) GetByIPAddress(ipAddress string, port int) (*models.Breaker, error) {
	var breaker models.Breaker
	err := r.db.Where("ip_address = ? AND port = ?", ipAddress, port).First(&breaker).Error
	if err != nil {
		return nil, err
	}
	return &breaker, nil
}

// GetEnabledBreakers 获取所有启用的断路器
func (r *breakerRepository) GetEnabledBreakers() ([]*models.Breaker, error) {
	var breakers []*models.Breaker
	err := r.db.Where("is_enabled = ?", true).
		Order("id ASC").
		Find(&breakers).Error
	return breakers, err
}

// Create 创建断路器
func (r *breakerRepository) Create(breaker *models.Breaker, device *models.Device) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建设备记录
		if err := tx.Create(device).Error; err != nil {
			return err
		}
		
		// 设置设备ID
		breaker.DeviceID = device.ID
		
		// 创建断路器记录
		return tx.Create(breaker).Error
	})
}

// Update 更新断路器
func (r *breakerRepository) Update(breaker *models.Breaker) error {
	return r.db.Save(breaker).Error
}

// Delete 删除断路器
func (r *breakerRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 获取断路器信息
		var breaker models.Breaker
		if err := tx.First(&breaker, id).Error; err != nil {
			return err
		}
		
		// 删除相关的控制记录
		if err := tx.Where("breaker_id = ?", id).Delete(&models.BreakerControl{}).Error; err != nil {
			return err
		}
		
		// 删除相关的绑定记录
		if err := tx.Where("breaker_id = ?", id).Delete(&models.BreakerServerBinding{}).Error; err != nil {
			return err
		}
		
		// 删除断路器记录
		if err := tx.Delete(&breaker).Error; err != nil {
			return err
		}
		
		// 删除设备记录
		return tx.Delete(&models.Device{}, breaker.DeviceID).Error
	})
}

// CreateControl 创建控制记录
func (r *breakerRepository) CreateControl(control *models.BreakerControl) error {
	return r.db.Create(control).Error
}

// UpdateControl 更新控制记录
func (r *breakerRepository) UpdateControl(control *models.BreakerControl) error {
	return r.db.Save(control).Error
}

// GetControl 获取控制记录
func (r *breakerRepository) GetControl(controlID string) (*models.BreakerControl, error) {
	var control models.BreakerControl
	err := r.db.Where("control_id = ?", controlID).First(&control).Error
	return &control, err
}

// GetBindings 获取断路器的绑定关系
func (r *breakerRepository) GetBindings(breakerID uint) ([]models.BreakerServerBinding, error) {
	var bindings []models.BreakerServerBinding
	err := r.db.Preload("Server").
		Where("breaker_id = ?", breakerID).
		Find(&bindings).Error
	return bindings, err
}

// CreateBinding 创建绑定关系
func (r *breakerRepository) CreateBinding(binding *models.BreakerServerBinding) error {
	return r.db.Create(binding).Error
}

// UpdateBinding 更新绑定关系
func (r *breakerRepository) UpdateBinding(binding *models.BreakerServerBinding) error {
	return r.db.Save(binding).Error
}

// DeleteBinding 删除绑定关系
func (r *breakerRepository) DeleteBinding(id uint) error {
	return r.db.Delete(&models.BreakerServerBinding{}, id).Error
}

// GetBinding 获取绑定关系
func (r *breakerRepository) GetBinding(id uint) (*models.BreakerServerBinding, error) {
	var binding models.BreakerServerBinding
	err := r.db.Preload("Breaker").
		Preload("Server").
		First(&binding, id).Error
	return &binding, err
}

// UpdateBreakerLockStatus 更新断路器锁定状态
func (r *breakerRepository) UpdateBreakerLockStatus(id uint, isLocked bool) error {
	return r.db.Model(&models.Breaker{}).Where("id = ?", id).Update("is_locked", isLocked).Error
}

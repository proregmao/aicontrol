package repositories

import (
	"gorm.io/gorm"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"
)

// AIStrategyRepository AI策略仓储接口
type AIStrategyRepository interface {
	// 策略CRUD操作
	FindStrategyByID(id uint) (*models.AIStrategy, error)
	FindAllStrategies() ([]models.AIStrategy, error)
	FindStrategiesByStatus(status models.AIStrategyStatus) ([]models.AIStrategy, error)
	FindEnabledStrategies() ([]*models.AIStrategy, error)
	FindStrategiesList(page, pageSize int, filters map[string]interface{}) ([]models.AIStrategy, int64, error)
	CreateStrategy(strategy *models.AIStrategy) error
	UpdateStrategy(strategy *models.AIStrategy) error
	DeleteStrategyByID(id uint) error
	
	// 策略执行记录操作
	CreateExecution(execution *models.AIStrategyExecution) error
	UpdateExecution(execution *models.AIStrategyExecution) error
	FindExecutionsByStrategyID(strategyID uint, page, pageSize int) ([]models.AIStrategyExecution, int64, error)
	FindAllExecutions(page, pageSize int) ([]models.AIStrategyExecution, int64, error)
	FindExecutionByID(id uint) (*models.AIStrategyExecution, error)
}

// aiStrategyRepository AI策略仓储实现
type aiStrategyRepository struct {
	db *gorm.DB
}

// NewAIStrategyRepository 创建AI策略仓储实例
func NewAIStrategyRepository() AIStrategyRepository {
	return &aiStrategyRepository{
		db: database.GetDB(),
	}
}

// FindStrategyByID 根据ID查找策略
func (r *aiStrategyRepository) FindStrategyByID(id uint) (*models.AIStrategy, error) {
	var strategy models.AIStrategy
	err := r.db.First(&strategy, id).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// FindAllStrategies 查找所有策略
func (r *aiStrategyRepository) FindAllStrategies() ([]models.AIStrategy, error) {
	var strategies []models.AIStrategy
	err := r.db.Order("priority DESC, created_at DESC").Find(&strategies).Error
	return strategies, err
}

// FindStrategiesByStatus 根据状态查找策略
func (r *aiStrategyRepository) FindStrategiesByStatus(status models.AIStrategyStatus) ([]models.AIStrategy, error) {
	var strategies []models.AIStrategy
	err := r.db.Where("status = ?", status).
		Order("priority DESC, created_at DESC").
		Find(&strategies).Error
	return strategies, err
}

// FindEnabledStrategies 查找所有启用的策略
func (r *aiStrategyRepository) FindEnabledStrategies() ([]*models.AIStrategy, error) {
	var strategies []*models.AIStrategy
	err := r.db.Where("status = ?", models.StrategyStatusEnabled).
		Order("priority DESC, created_at DESC").
		Find(&strategies).Error
	return strategies, err
}

// FindStrategiesList 分页查询策略列表
func (r *aiStrategyRepository) FindStrategiesList(page, pageSize int, filters map[string]interface{}) ([]models.AIStrategy, int64, error) {
	var strategies []models.AIStrategy
	var total int64

	query := r.db.Model(&models.AIStrategy{})

	// 应用过滤条件
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name.(string)+"%")
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if createdBy, ok := filters["created_by"]; ok && createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Scopes(database.Paginate(page, pageSize)).
		Order("priority DESC, created_at DESC").
		Find(&strategies).Error

	return strategies, total, err
}

// CreateStrategy 创建策略
func (r *aiStrategyRepository) CreateStrategy(strategy *models.AIStrategy) error {
	return r.db.Create(strategy).Error
}

// UpdateStrategy 更新策略
func (r *aiStrategyRepository) UpdateStrategy(strategy *models.AIStrategy) error {
	return r.db.Save(strategy).Error
}

// DeleteStrategyByID 根据ID删除策略（软删除）
func (r *aiStrategyRepository) DeleteStrategyByID(id uint) error {
	return r.db.Delete(&models.AIStrategy{}, id).Error
}

// CreateExecution 创建执行记录
func (r *aiStrategyRepository) CreateExecution(execution *models.AIStrategyExecution) error {
	return r.db.Create(execution).Error
}

// UpdateExecution 更新执行记录
func (r *aiStrategyRepository) UpdateExecution(execution *models.AIStrategyExecution) error {
	return r.db.Save(execution).Error
}

// FindExecutionsByStrategyID 根据策略ID查找执行记录
func (r *aiStrategyRepository) FindExecutionsByStrategyID(strategyID uint, page, pageSize int) ([]models.AIStrategyExecution, int64, error) {
	var executions []models.AIStrategyExecution
	var total int64

	query := r.db.Model(&models.AIStrategyExecution{}).Where("strategy_id = ?", strategyID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载策略信息
	err := query.Scopes(database.Paginate(page, pageSize)).
		Preload("Strategy").
		Order("executed_at DESC").
		Find(&executions).Error

	return executions, total, err
}

// FindAllExecutions 查找所有执行记录
func (r *aiStrategyRepository) FindAllExecutions(page, pageSize int) ([]models.AIStrategyExecution, int64, error) {
	var executions []models.AIStrategyExecution
	var total int64

	query := r.db.Model(&models.AIStrategyExecution{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载策略信息
	err := query.Scopes(database.Paginate(page, pageSize)).
		Preload("Strategy").
		Order("executed_at DESC").
		Find(&executions).Error

	return executions, total, err
}

// FindExecutionByID 根据ID查找执行记录
func (r *aiStrategyRepository) FindExecutionByID(id uint) (*models.AIStrategyExecution, error) {
	var execution models.AIStrategyExecution
	err := r.db.Preload("Strategy").First(&execution, id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

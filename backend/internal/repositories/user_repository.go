package repositories

import (
	"gorm.io/gorm"
	"smart-device-management/internal/models"
	"smart-device-management/pkg/database"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	FindUserByID(id uint) (*models.User, error)
	FindUserByUsername(username string) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUserByID(id uint) error
	FindUsersByStatus(status models.UserStatus) ([]models.User, error)
	FindUsersList(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error)
	ExistsUserByUsername(username string) (bool, error)
	ExistsUserByEmail(email string) (bool, error)
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.GetDB(),
	}
}

// FindUserByID 根据ID查找用户
func (r *userRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByUsername 根据用户名查找用户
func (r *userRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail 根据邮箱查找用户
func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// UpdateUser 更新用户
func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// DeleteUserByID 根据ID删除用户（软删除）
func (r *userRepository) DeleteUserByID(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// FindUsersByStatus 根据状态查找用户
func (r *userRepository) FindUsersByStatus(status models.UserStatus) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("status = ?", status).Find(&users).Error
	return users, err
}

// FindUsersList 分页查询用户列表
func (r *userRepository) FindUsersList(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// 应用过滤条件
	if username, ok := filters["username"]; ok && username != "" {
		query = query.Where("username ILIKE ?", "%"+username.(string)+"%")
	}
	if email, ok := filters["email"]; ok && email != "" {
		query = query.Where("email ILIKE ?", "%"+email.(string)+"%")
	}
	if role, ok := filters["role"]; ok && role != "" {
		query = query.Where("role = ?", role)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Scopes(database.Paginate(page, pageSize)).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

// ExistsUserByUsername 检查用户名是否存在
func (r *userRepository) ExistsUserByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsUserByEmail 检查邮箱是否存在
func (r *userRepository) ExistsUserByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

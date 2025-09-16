package services

import (
	"errors"
	"time"

	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/internal/shared"
	"smart-device-management/internal/utils"
	"smart-device-management/pkg/database"
	"smart-device-management/pkg/security"
)

// UserService 用户服务接口
type UserService interface {
	Login(request *models.LoginRequest) (*models.LoginResponse, error)
	Logout(tokenString string) error
	CreateUser(request *models.CreateUserRequest) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserList(page, pageSize int, filters map[string]interface{}) (*models.UserListResponse, error)
	UpdateUser(id uint, request *models.UpdateUserRequest) (*models.User, error)
	DeleteUser(id uint) error
	ChangePassword(userID uint, request *models.ChangePasswordRequest) error
	RefreshToken(refreshToken string) (*models.LoginResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repositories.UserRepository
	jwtManager *security.JWTManager
}

// NewUserService 创建用户服务实例
func NewUserService() UserService {
	return &userService{
		userRepo:   repositories.NewUserRepository(),
		jwtManager: shared.GlobalJWTManager,
	}
}

// Login 用户登录
func (s *userService) Login(request *models.LoginRequest) (*models.LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.FindUserByUsername(request.Username)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("用户账户已被禁用")
	}

	// 验证密码
	if !utils.CheckPassword(request.Password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成Token
	token, err := utils.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.UpdateUser(user)

	return &models.LoginResponse{
		Token:     token,
		ExpiresIn: utils.GetTokenExpirationTime(),
		User:      *user,
	}, nil
}

// Logout 用户登出
func (s *userService) Logout(tokenString string) error {
	// 将Token加入黑名单
	s.jwtManager.BlacklistToken(tokenString)
	return nil
}

// CreateUser 创建用户
func (s *userService) CreateUser(request *models.CreateUserRequest) (*models.User, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsUserByUsername(request.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if request.Email != "" {
		exists, err := s.userRepo.ExistsUserByEmail(request.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 验证密码强度
	if !utils.ValidatePasswordStrength(request.Password) {
		return nil, errors.New("密码强度不足")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     request.Username,
		PasswordHash: hashedPassword,
		Email:        request.Email,
		FullName:     request.FullName,
		Role:         request.Role,
		Status:       models.StatusActive,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(id)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return user, nil
}

// GetUserList 获取用户列表
func (s *userService) GetUserList(page, pageSize int, filters map[string]interface{}) (*models.UserListResponse, error) {
	users, total, err := s.userRepo.FindUsersList(page, pageSize, filters)
	if err != nil {
		return nil, err
	}

	return &models.UserListResponse{
		Users: users,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(id uint, request *models.UpdateUserRequest) (*models.User, error) {
	// 查找用户
	user, err := s.userRepo.FindUserByID(id)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查邮箱是否已被其他用户使用
	if request.Email != "" && request.Email != user.Email {
		exists, err := s.userRepo.ExistsUserByEmail(request.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已被其他用户使用")
		}
	}

	// 更新字段
	if request.Email != "" {
		user.Email = request.Email
	}
	if request.FullName != "" {
		user.FullName = request.FullName
	}
	if request.Role != "" {
		user.Role = request.Role
	}
	if request.Status != "" {
		user.Status = request.Status
	}

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.FindUserByID(id)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return errors.New("用户不存在")
		}
		return err
	}

	return s.userRepo.DeleteUserByID(id)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, request *models.ChangePasswordRequest) error {
	// 查找用户
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(request.OldPassword, user.PasswordHash) {
		return errors.New("原密码错误")
	}

	// 验证新密码强度
	if !utils.ValidatePasswordStrength(request.NewPassword) {
		return errors.New("新密码强度不足")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	return s.userRepo.UpdateUser(user)
}

// RefreshToken 刷新Token
func (s *userService) RefreshToken(currentToken string) (*models.LoginResponse, error) {
	// 解析当前Token
	claims, err := utils.ParseToken(currentToken)
	if err != nil {
		return nil, errors.New("无效的Token")
	}

	// 查找用户
	user, err := s.userRepo.FindUserByID(claims.UserID)
	if err != nil {
		if database.IsRecordNotFoundError(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("用户账户已被禁用")
	}

	// 将旧Token加入黑名单
	s.jwtManager.BlacklistToken(currentToken)

	// 生成新Token
	token, err := utils.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresIn: utils.GetTokenExpirationTime(),
		User:      *user,
	}, nil
}

package services

import (
	"errors"
	"fmt"
	"smart-device-management/internal/models"
	"smart-device-management/internal/repositories"
	"smart-device-management/pkg/logger"
	"strconv"
	"time"
)

type DeviceService struct {
	deviceRepo *repositories.DeviceRepository
	logger     *logger.Logger
}

func NewDeviceService(deviceRepo *repositories.DeviceRepository, logger *logger.Logger) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

// GetAllDevices 获取所有设备
func (s *DeviceService) GetAllDevices() ([]models.Device, error) {
	s.logger.Info("获取所有设备列表")

	devices, err := s.deviceRepo.FindAll()
	if err != nil {
		s.logger.Error("获取设备列表失败", "error", err)
		return nil, fmt.Errorf("获取设备列表失败: %w", err)
	}

	s.logger.Info("成功获取设备列表", "count", len(devices))
	return devices, nil
}

// GetDeviceByID 根据ID获取设备
func (s *DeviceService) GetDeviceByID(id string) (*models.Device, error) {
	s.logger.Info("获取设备详情", "device_id", id)

	// 将字符串ID转换为uint
	deviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的设备ID格式", "device_id", id, "error", err)
		return nil, errors.New("无效的设备ID格式")
	}

	device, err := s.deviceRepo.FindByID(uint(deviceID))
	if err != nil {
		s.logger.Error("获取设备详情失败", "device_id", id, "error", err)
		return nil, fmt.Errorf("获取设备详情失败: %w", err)
	}

	if device == nil {
		s.logger.Warn("设备不存在", "device_id", id)
		return nil, errors.New("设备不存在")
	}

	s.logger.Info("成功获取设备详情", "device_id", id, "device_name", device.DeviceName)
	return device, nil
}

// GetDevicesByType 根据类型获取设备
func (s *DeviceService) GetDevicesByType(deviceType string) ([]models.Device, error) {
	s.logger.Info("根据类型获取设备", "device_type", deviceType)

	// 验证设备类型
	validTypes := []string{"temperature_sensor", "breaker", "server"}
	isValid := false
	for _, validType := range validTypes {
		if deviceType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		s.logger.Error("无效的设备类型", "device_type", deviceType)
		return nil, errors.New("无效的设备类型")
	}

	devices, err := s.deviceRepo.FindByType(deviceType)
	if err != nil {
		s.logger.Error("根据类型获取设备失败", "device_type", deviceType, "error", err)
		return nil, fmt.Errorf("根据类型获取设备失败: %w", err)
	}

	s.logger.Info("成功根据类型获取设备", "device_type", deviceType, "count", len(devices))
	return devices, nil
}

// CreateDevice 创建设备
func (s *DeviceService) CreateDevice(req *models.CreateDeviceRequest) (*models.Device, error) {
	s.logger.Info("创建设备", "device_name", req.DeviceName, "device_type", req.DeviceType)

	// 验证设备类型
	validTypes := []models.DeviceType{models.DeviceTypeTemperatureSensor, models.DeviceTypeBreaker, models.DeviceTypeServer}
	isValid := false
	for _, validType := range validTypes {
		if req.DeviceType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		s.logger.Error("无效的设备类型", "device_type", req.DeviceType)
		return nil, errors.New("无效的设备类型")
	}

	// 检查设备名称是否已存在
	existingDevice, err := s.deviceRepo.FindByName(req.DeviceName)
	if err != nil {
		s.logger.Error("检查设备名称失败", "device_name", req.DeviceName, "error", err)
		return nil, fmt.Errorf("检查设备名称失败: %w", err)
	}

	if existingDevice != nil {
		s.logger.Error("设备名称已存在", "device_name", req.DeviceName)
		return nil, errors.New("设备名称已存在")
	}

	// 创建设备对象
	device := &models.Device{
		DeviceType:  req.DeviceType,
		DeviceName:  req.DeviceName,
		DeviceModel: req.DeviceModel,
		Location:    req.Location,
		IPAddress:   req.IPAddress,
		Port:        req.Port,
		SlaveID:     req.SlaveID,
		Description: req.Description,
		Config:      req.Config,
		Status:      models.DeviceStatusOnline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存设备
	err = s.deviceRepo.Create(device)
	if err != nil {
		s.logger.Error("创建设备失败", "device_name", req.DeviceName, "error", err)
		return nil, fmt.Errorf("创建设备失败: %w", err)
	}

	s.logger.Info("成功创建设备", "device_id", device.ID, "device_name", device.DeviceName)
	return device, nil
}

// UpdateDevice 更新设备
func (s *DeviceService) UpdateDevice(id string, req *models.UpdateDeviceRequest) (*models.Device, error) {
	s.logger.Info("更新设备", "device_id", id)

	// 将字符串ID转换为uint
	deviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的设备ID格式", "device_id", id, "error", err)
		return nil, errors.New("无效的设备ID格式")
	}

	// 获取现有设备
	device, err := s.deviceRepo.FindByID(uint(deviceID))
	if err != nil {
		s.logger.Error("获取设备失败", "device_id", id, "error", err)
		return nil, fmt.Errorf("获取设备失败: %w", err)
	}

	if device == nil {
		s.logger.Warn("设备不存在", "device_id", id)
		return nil, errors.New("设备不存在")
	}

	// 更新设备信息
	if req.DeviceName != "" {
		device.DeviceName = req.DeviceName
	}
	if req.DeviceModel != "" {
		device.DeviceModel = req.DeviceModel
	}
	if req.Location != "" {
		device.Location = req.Location
	}
	if req.IPAddress != "" {
		device.IPAddress = req.IPAddress
	}
	if req.Port != 0 {
		device.Port = req.Port
	}
	if req.SlaveID != 0 {
		device.SlaveID = req.SlaveID
	}
	if req.Description != "" {
		device.Description = req.Description
	}
	if req.Config != nil {
		device.Config = req.Config
	}
	if req.Status != "" {
		device.Status = req.Status
	}
	device.UpdatedAt = time.Now()

	// 保存更新
	err = s.deviceRepo.Update(device)
	if err != nil {
		s.logger.Error("更新设备失败", "device_id", id, "error", err)
		return nil, fmt.Errorf("更新设备失败: %w", err)
	}

	s.logger.Info("成功更新设备", "device_id", id, "device_name", device.DeviceName)
	return device, nil
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(id string) error {
	s.logger.Info("删除设备", "device_id", id)

	// 将字符串ID转换为uint
	deviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的设备ID格式", "device_id", id, "error", err)
		return errors.New("无效的设备ID格式")
	}

	// 检查设备是否存在
	device, err := s.deviceRepo.FindByID(uint(deviceID))
	if err != nil {
		s.logger.Error("获取设备失败", "device_id", id, "error", err)
		return fmt.Errorf("获取设备失败: %w", err)
	}

	if device == nil {
		s.logger.Warn("设备不存在", "device_id", id)
		return errors.New("设备不存在")
	}

	// 删除设备
	err = s.deviceRepo.Delete(uint(deviceID))
	if err != nil {
		s.logger.Error("删除设备失败", "device_id", id, "error", err)
		return fmt.Errorf("删除设备失败: %w", err)
	}

	s.logger.Info("成功删除设备", "device_id", id, "device_name", device.DeviceName)
	return nil
}

// UpdateDeviceStatus 更新设备状态
func (s *DeviceService) UpdateDeviceStatus(id string, status string) error {
	s.logger.Info("更新设备状态", "device_id", id, "status", status)

	// 验证状态值
	validStatuses := []string{"online", "offline", "error", "maintenance"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		s.logger.Error("无效的设备状态", "status", status)
		return errors.New("无效的设备状态")
	}

	// 将字符串ID转换为uint
	deviceID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		s.logger.Error("无效的设备ID格式", "device_id", id, "error", err)
		return errors.New("无效的设备ID格式")
	}

	// 更新设备状态
	err = s.deviceRepo.UpdateStatus(uint(deviceID), status)
	if err != nil {
		s.logger.Error("更新设备状态失败", "device_id", id, "status", status, "error", err)
		return fmt.Errorf("更新设备状态失败: %w", err)
	}

	s.logger.Info("成功更新设备状态", "device_id", id, "status", status)
	return nil
}

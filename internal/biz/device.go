package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// Device 设备实体
type Device struct {
	ID              string
	UserID          int64
	MacAddress      string
	LastConnectedAt *time.Time
	AutoUpdate      int32
	Board           string
	Alias           string
	AgentID         string
	AppVersion      string
	Sort            int32
	Creator         int64
	CreateDate      time.Time
	Updater         int64
	UpdateDate      time.Time
}

// DeviceRepo 设备数据访问接口
type DeviceRepo interface {
	Create(ctx context.Context, device *Device) error
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, deviceId string, userId int64) error
	GetByID(ctx context.Context, deviceId string) (*Device, error)
	GetByMacAddress(ctx context.Context, macAddress string) (*Device, error)
	ListByUserAndAgent(ctx context.Context, userId int64, agentId string) ([]*Device, error)
	SelectCountByUserId(ctx context.Context, userId int64) (int64, error)
	DeleteByUserId(ctx context.Context, userId int64) error
	PageDevices(ctx context.Context, keywords *string, page *kit.PageRequest) ([]*Device, error)
	TotalDevices(ctx context.Context, keywords *string) (int, error)
}

// DeviceUsecase 设备业务逻辑
type DeviceUsecase struct {
	repo        DeviceRepo
	redisClient *kit.RedisClient
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewDeviceUsecase 创建设备用例
func NewDeviceUsecase(
	repo DeviceRepo,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *DeviceUsecase {
	return &DeviceUsecase{
		repo:        repo,
		redisClient: redisClient,
		handleError: cerrors.NewHandleError(logger),
		log:         log.NewHelper(log.With(logger, "module", "agent-matrix-service/biz/device")),
	}
}

// DeviceActivation 设备激活（完整实现）
func (uc *DeviceUsecase) DeviceActivation(ctx context.Context, userId int64, agentId, activationCode string, deviceId, macAddress, board, appVersion string) error {
	if activationCode == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("激活码不能为空"))
	}

	// 检查设备是否已被激活
	existingDevice, err := uc.repo.GetByID(ctx, deviceId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existingDevice != nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("设备已被激活"))
	}

	// 创建设备
	now := time.Now()
	device := &Device{
		ID:              deviceId,
		UserID:          userId,
		MacAddress:      macAddress,
		LastConnectedAt: &now,
		AutoUpdate:      1,
		Board:           board,
		AgentID:         agentId,
		AppVersion:      appVersion,
		Creator:         userId,
		CreateDate:      now,
		Updater:         userId,
		UpdateDate:      now,
	}

	err = uc.repo.Create(ctx, device)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	return nil
}

// GetUserDevices 获取用户设备列表
func (uc *DeviceUsecase) GetUserDevices(ctx context.Context, userId int64, agentId string) ([]*Device, error) {
	devices, err := uc.repo.ListByUserAndAgent(ctx, userId, agentId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	return devices, nil
}

// UnbindDevice 解绑设备
func (uc *DeviceUsecase) UnbindDevice(ctx context.Context, userId int64, deviceId string) error {
	// 先查询设备，确保设备存在且属于该用户
	device, err := uc.repo.GetByID(ctx, deviceId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if device == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("设备不存在"))
	}
	if device.UserID != userId {
		return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("无权操作该设备"))
	}

	err = uc.repo.Delete(ctx, deviceId, userId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	return nil
}

// UpdateDevice 更新设备信息
func (uc *DeviceUsecase) UpdateDevice(ctx context.Context, deviceId string, userId int64, autoUpdate *int32, alias *string) error {
	// 先查询设备，确保设备存在且属于该用户
	device, err := uc.repo.GetByID(ctx, deviceId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if device == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("设备不存在"))
	}
	if device.UserID != userId {
		return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("无权操作该设备"))
	}

	// 更新字段
	if autoUpdate != nil {
		device.AutoUpdate = *autoUpdate
	}
	if alias != nil {
		device.Alias = *alias
	}
	device.Updater = userId
	device.UpdateDate = time.Now()

	err = uc.repo.Update(ctx, device)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	return nil
}

// ManualAddDevice 手动添加设备
func (uc *DeviceUsecase) ManualAddDevice(ctx context.Context, userId int64, agentId, board, appVersion, macAddress string) error {
	// 检查MAC地址是否已存在
	existingDevice, err := uc.repo.GetByMacAddress(ctx, macAddress)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existingDevice != nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("MAC地址已存在"))
	}

	now := time.Now()
	device := &Device{
		ID:              macAddress, // 使用MAC地址作为ID
		UserID:          userId,
		AgentID:         agentId,
		Board:           board,
		AppVersion:      appVersion,
		MacAddress:      macAddress,
		CreateDate:      now,
		UpdateDate:      now,
		LastConnectedAt: &now,
		Creator:         userId,
		Updater:         userId,
		AutoUpdate:      1,
	}

	err = uc.repo.Create(ctx, device)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	return nil
}

// GetDeviceByID 根据ID获取设备
func (uc *DeviceUsecase) GetDeviceByID(ctx context.Context, deviceId string) (*Device, error) {
	device, err := uc.repo.GetByID(ctx, deviceId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	return device, nil
}

// GetDeviceByMacAddress 根据MAC地址获取设备
func (uc *DeviceUsecase) GetDeviceByMacAddress(ctx context.Context, macAddress string) (*Device, error) {
	device, err := uc.repo.GetByMacAddress(ctx, macAddress)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	return device, nil
}

// GeCodeByDeviceId 从 Redis 获取设备激活码
// 对应 Java 的 deviceService.geCodeByDeviceId 方法
func (uc *DeviceUsecase) GeCodeByDeviceId(ctx context.Context, deviceId string) (string, error) {
	if deviceId == "" {
		return "", nil
	}

	// 构建 Redis Key: ota:activation:data:{deviceId}
	// deviceId 需要替换 : 为 _ 并转小写
	safeDeviceId := strings.ToLower(strings.ReplaceAll(deviceId, ":", "_"))
	redisKey := fmt.Sprintf("ota:activation:data:%s", safeDeviceId)

	// 从 Redis 获取 Map
	var cacheMap map[string]interface{}
	err := uc.redisClient.GetObject(ctx, redisKey, &cacheMap)
	if err != nil {
		// Redis 中没有数据，返回空字符串（不是错误）
		return "", nil
	}

	// 获取 activation_code 字段
	if cacheMap != nil {
		if code, ok := cacheMap["activation_code"].(string); ok {
			return code, nil
		}
	}

	return "", nil
}

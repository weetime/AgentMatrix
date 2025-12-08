package biz

import (
	"context"
	"fmt"
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
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewDeviceUsecase 创建设备用例
func NewDeviceUsecase(
	repo DeviceRepo,
	logger log.Logger,
) *DeviceUsecase {
	return &DeviceUsecase{
		repo:        repo,
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

// AdminPageDeviceVO 管理员分页设备VO
type AdminPageDeviceVO struct {
	ID             string `json:"id"`
	MacAddress     string `json:"macAddress"`
	BindUserName   string `json:"bindUserName"`
	DeviceType     string `json:"deviceType"`
	AppVersion     string `json:"appVersion"`
	OtaUpgrade     int32  `json:"otaUpgrade"`
	RecentChatTime string `json:"recentChatTime"`
}

// AdminUserRepo 用户数据访问接口（用于查询用户名）
type AdminUserRepo interface {
	GetUsersByIDs(ctx context.Context, userIds []int64) (map[int64]*AdminUser, error)
}

// AdminUser 用户实体（用于查询用户名）
type AdminUser struct {
	ID       int64
	Username string
}

// PageAdminDevices 管理员分页查询设备
func (uc *DeviceUsecase) PageAdminDevices(ctx context.Context, keywords *string, page *kit.PageRequest, userRepo AdminUserRepo) ([]*AdminPageDeviceVO, int, error) {
	// 查询设备列表
	devices, err := uc.repo.PageDevices(ctx, keywords, page)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	// 查询总数
	total, err := uc.repo.TotalDevices(ctx, keywords)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	// 批量查询用户名
	userIds := make([]int64, 0)
	userIdSet := make(map[int64]bool)
	for _, device := range devices {
		if !userIdSet[device.UserID] {
			userIds = append(userIds, device.UserID)
			userIdSet[device.UserID] = true
		}
	}

	userMap := make(map[int64]string)
	if len(userIds) > 0 && userRepo != nil {
		users, err := userRepo.GetUsersByIDs(ctx, userIds)
		if err != nil {
			uc.log.Warnf("Failed to get users: %v", err)
		} else {
			for userId, user := range users {
				userMap[userId] = user.Username
			}
		}
	}

	// 转换为VO
	voList := make([]*AdminPageDeviceVO, len(devices))
	for i, device := range devices {
		recentChatTime := ""
		if device.UpdateDate.After(device.CreateDate) {
			recentChatTime = formatShortTime(device.UpdateDate)
		}

		voList[i] = &AdminPageDeviceVO{
			ID:             device.ID,
			MacAddress:     device.MacAddress,
			BindUserName:   userMap[device.UserID],
			DeviceType:     device.Board,
			AppVersion:     device.AppVersion,
			OtaUpgrade:     device.AutoUpdate,
			RecentChatTime: recentChatTime,
		}
	}

	return voList, total, nil
}

// formatShortTime 格式化时间为简短描述
func formatShortTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	}

	return t.Format("2006-01-02 15:04:05")
}

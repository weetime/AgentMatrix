package biz

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrDuplicateOtaTypeVersion = fmt.Errorf("已存在相同类型和版本的固件，请修改后重试")
)

// Ota OTA固件实体
type Ota struct {
	ID           string    // ID
	FirmwareName string    // 固件名称
	Type         string    // 固件类型
	Version      string    // 版本号
	Size         int64     // 文件大小(字节)
	Remark       string    // 备注/说明
	FirmwarePath string    // 固件路径
	Sort         int32     // 排序
	Updater      int64     // 更新者ID
	UpdateDate   time.Time // 更新时间
	Creator      int64     // 创建者ID
	CreateDate   time.Time // 创建时间
}

// OtaResponseDTO OTA固件响应DTO（用于向前端展示，所有int64字段格式化为字符串）
type OtaResponseDTO struct {
	ID           string // ID
	FirmwareName string // 固件名称
	Type         string // 固件类型
	Version      string // 版本号
	Size         int64  // 文件大小(字节)
	Remark       string // 备注/说明
	FirmwarePath string // 固件路径
	Sort         int32  // 排序
	Updater      string // 更新者ID（格式化为字符串）
	UpdateDate   string // 更新时间（格式化）
	Creator      string // 创建者ID（格式化为字符串）
	CreateDate   string // 创建时间（格式化）
}

// ListOtaParams 分页查询参数
type ListOtaParams struct {
	FirmwareName *string // 可选，固件名称（模糊查询）
}

// OtaRepo OTA固件数据访问接口
type OtaRepo interface {
	// PageOta 分页查询OTA固件列表
	PageOta(ctx context.Context, params *ListOtaParams, page *kit.PageRequest) ([]*Ota, int, error)

	// GetByID 根据ID查询OTA固件记录
	GetByID(ctx context.Context, id string) (*Ota, error)

	// Save 保存OTA固件（同类固件只保留最新一条）
	Save(ctx context.Context, entity *Ota) error

	// Update 更新OTA固件（检查类型和版本唯一性）
	Update(ctx context.Context, entity *Ota) error

	// Delete 批量删除OTA固件
	Delete(ctx context.Context, ids []string) error

	// GetLatestOta 根据类型获取最新OTA固件
	GetLatestOta(ctx context.Context, otaType string) (*Ota, error)
}

// DeviceReportReqDTO 设备上报请求DTO
type DeviceReportReqDTO struct {
	Version             *int32       `json:"version,omitempty"`
	FlashSize           *int32       `json:"flash_size,omitempty"`
	MinimumFreeHeapSize *int32       `json:"minimum_free_heap_size,omitempty"`
	MacAddress          *string      `json:"mac_address,omitempty"`
	UUID                *string      `json:"uuid,omitempty"`
	ChipModelName       *string      `json:"chip_model_name,omitempty"`
	ChipInfo            *ChipInfo    `json:"chip_info,omitempty"`
	Application         *Application `json:"application,omitempty"`
	PartitionTable      []*Partition `json:"partition_table,omitempty"`
	Ota                 *OtaInfo     `json:"ota,omitempty"`
	Board               *BoardInfo   `json:"board,omitempty"`
}

type ChipInfo struct {
	Model    *int32 `json:"model,omitempty"`
	Cores    *int32 `json:"cores,omitempty"`
	Revision *int32 `json:"revision,omitempty"`
	Features *int32 `json:"features,omitempty"`
}

type Application struct {
	Name        *string `json:"name,omitempty"`
	Version     *string `json:"version,omitempty"`
	CompileTime *string `json:"compile_time,omitempty"`
	IdfVersion  *string `json:"idf_version,omitempty"`
	ElfSha256   *string `json:"elf_sha256,omitempty"`
}

type Partition struct {
	Label   *string `json:"label,omitempty"`
	Type    *int32  `json:"type,omitempty"`
	Subtype *int32  `json:"subtype,omitempty"`
	Address *int32  `json:"address,omitempty"`
	Size    *int32  `json:"size,omitempty"`
}

type OtaInfo struct {
	Label *string `json:"label,omitempty"`
}

type BoardInfo struct {
	Type    *string `json:"type,omitempty"`
	SSID    *string `json:"ssid,omitempty"`
	RSSI    *int32  `json:"rssi,omitempty"`
	Channel *int32  `json:"channel,omitempty"`
	IP      *string `json:"ip,omitempty"`
	MAC     *string `json:"mac,omitempty"`
}

// DeviceReportRespDTO 设备上报响应DTO
type DeviceReportRespDTO struct {
	ServerTime *ServerTime `json:"server_time,omitempty"`
	Activation *Activation `json:"activation,omitempty"`
	Error      string      `json:"error,omitempty"`
	Firmware   *Firmware   `json:"firmware,omitempty"`
	Websocket  *Websocket  `json:"websocket,omitempty"`
	MQTT       *MQTT       `json:"mqtt,omitempty"`
}

type ServerTime struct {
	Timestamp      int64  `json:"timestamp"`
	TimeZone       string `json:"timeZone"`
	TimezoneOffset int32  `json:"timezone_offset"`
}

type Activation struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Challenge string `json:"challenge"`
}

type Firmware struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

type Websocket struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type MQTT struct {
	Endpoint       string `json:"endpoint"`
	ClientID       string `json:"client_id"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PublishTopic   string `json:"publish_topic"`
	SubscribeTopic string `json:"subscribe_topic"`
}

// OtaUsecase OTA固件业务逻辑
type OtaUsecase struct {
	repo          OtaRepo
	DeviceUsecase *DeviceUsecase
	ConfigUsecase *ConfigUsecase
	redisClient   *kit.RedisClient
	handleError   *cerrors.HandleError
	log           *log.Helper
}

// NewOtaUsecase 创建OTA固件用例
func NewOtaUsecase(
	repo OtaRepo,
	deviceUsecase *DeviceUsecase,
	configUsecase *ConfigUsecase,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *OtaUsecase {
	return &OtaUsecase{
		repo:          repo,
		DeviceUsecase: deviceUsecase,
		ConfigUsecase: configUsecase,
		redisClient:   redisClient,
		handleError:   cerrors.NewHandleError(logger),
		log:           kit.LogHelper(logger),
	}
}

// PageOta 分页查询OTA固件列表
func (uc *OtaUsecase) PageOta(ctx context.Context, params *ListOtaParams, page *kit.PageRequest) ([]*OtaResponseDTO, int, error) {
	// 查询列表
	list, total, err := uc.repo.PageOta(ctx, params, page)
	if err != nil {
		return nil, 0, err
	}

	// 转换为DTO列表
	dtoList := make([]*OtaResponseDTO, 0, len(list))
	for _, entity := range list {
		dto := uc.convertToResponseDTO(entity)
		dtoList = append(dtoList, dto)
	}

	return dtoList, total, nil
}

// GetOtaByID 根据ID查询OTA固件信息
func (uc *OtaUsecase) GetOtaByID(ctx context.Context, id string) (*OtaResponseDTO, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("OTA固件记录不存在"))
	}

	return uc.convertToResponseDTO(entity), nil
}

// SaveOta 保存OTA固件
func (uc *OtaUsecase) SaveOta(ctx context.Context, entity *Ota) error {
	if entity == nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件信息不能为空"))
	}
	if entity.FirmwareName == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件名称不能为空"))
	}
	if entity.Type == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件类型不能为空"))
	}
	if entity.Version == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("版本号不能为空"))
	}

	if err := uc.repo.Save(ctx, entity); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// UpdateOta 更新OTA固件
func (uc *OtaUsecase) UpdateOta(ctx context.Context, id string, entity *Ota) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if entity == nil {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件信息不能为空"))
	}
	if entity.FirmwareName == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件名称不能为空"))
	}
	if entity.Type == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("固件类型不能为空"))
	}
	if entity.Version == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("版本号不能为空"))
	}

	entity.ID = id
	if err := uc.repo.Update(ctx, entity); err != nil {
		if err == ErrDuplicateOtaTypeVersion {
			return err
		}
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// DeleteOta 批量删除OTA固件
func (uc *OtaUsecase) DeleteOta(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("删除的固件ID不能为空"))
	}

	if err := uc.repo.Delete(ctx, ids); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// GetDownloadUrl 生成下载链接UUID并存储到Redis
func (uc *OtaUsecase) GetDownloadUrl(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	// 验证记录存在
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return "", uc.handleError.ErrInternal(ctx, err)
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 存储到Redis（24小时过期）
	redisKey := kit.GetOtaIdKey(uuidStr)
	if err := uc.redisClient.Set(ctx, redisKey, id, 24*time.Hour); err != nil {
		return "", uc.handleError.ErrInternal(ctx, err)
	}

	return uuidStr, nil
}

// DownloadOta 下载OTA固件（从Redis获取ID，检查下载次数，返回文件数据）
func (uc *OtaUsecase) DownloadOta(ctx context.Context, uuidStr string) ([]byte, string, error) {
	if uuidStr == "" {
		return nil, "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("uuid不能为空"))
	}

	// 从Redis获取ID
	redisKey := kit.GetOtaIdKey(uuidStr)
	id, err := uc.redisClient.Get(ctx, redisKey)
	if err != nil || id == "" {
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("下载链接不存在或已过期"))
	}

	// 检查下载次数
	downloadCountKey := kit.GetOtaDownloadCountKey(uuidStr)
	downloadCountStr, _ := uc.redisClient.Get(ctx, downloadCountKey)
	downloadCount := 0
	if downloadCountStr != "" {
		fmt.Sscanf(downloadCountStr, "%d", &downloadCount)
	}

	// 如果下载次数超过3次，删除Redis key并返回错误
	if downloadCount >= 3 {
		_ = uc.redisClient.Delete(ctx, downloadCountKey)
		_ = uc.redisClient.Delete(ctx, redisKey)
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("下载次数已达上限"))
	}

	// 增加下载次数
	downloadCount++
	if err := uc.redisClient.Set(ctx, downloadCountKey, fmt.Sprintf("%d", downloadCount), 24*time.Hour); err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}

	// 获取固件信息
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil || entity.FirmwarePath == "" {
		return nil, "", uc.handleError.ErrNotFound(ctx, fmt.Errorf("固件文件不存在"))
	}

	// 读取文件
	fileData, filename, err := uc.readFirmwareFile(entity.FirmwarePath, entity.Type, entity.Version)
	if err != nil {
		return nil, "", uc.handleError.ErrInternal(ctx, err)
	}

	return fileData, filename, nil
}

// UploadFirmware 上传固件文件（计算MD5，保存文件，返回路径）
func (uc *OtaUsecase) UploadFirmware(ctx context.Context, fileData []byte, filename string) (string, error) {
	if len(fileData) == 0 {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("上传文件不能为空"))
	}
	if filename == "" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("文件名不能为空"))
	}

	// 检查文件扩展名
	ext := filepath.Ext(filename)
	if ext != ".bin" && ext != ".apk" {
		return "", uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("只允许上传.bin和.apk格式的文件"))
	}

	// 计算MD5值
	md5Hash := md5.Sum(fileData)
	md5Str := fmt.Sprintf("%x", md5Hash)

	// 设置存储路径
	uploadDir := "uploadfile"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", uc.handleError.ErrInternal(ctx, fmt.Errorf("创建上传目录失败: %w", err))
	}

	// 使用MD5作为文件名，保留原始扩展名
	uniqueFileName := md5Str + ext
	filePath := filepath.Join(uploadDir, uniqueFileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件已存在，直接返回路径
		return filePath, nil
	}

	// 保存文件
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", uc.handleError.ErrInternal(ctx, fmt.Errorf("保存文件失败: %w", err))
	}

	return filePath, nil
}

// convertToResponseDTO 转换为响应DTO
func (uc *OtaUsecase) convertToResponseDTO(entity *Ota) *OtaResponseDTO {
	dto := &OtaResponseDTO{
		ID:           entity.ID,
		FirmwareName: entity.FirmwareName,
		Type:         entity.Type,
		Version:      entity.Version,
		Size:         entity.Size,
		Remark:       entity.Remark,
		FirmwarePath: entity.FirmwarePath,
		Sort:         entity.Sort,
	}

	// 格式化int64字段为字符串
	if entity.Creator > 0 {
		dto.Creator = fmt.Sprintf("%d", entity.Creator)
	}
	if entity.Updater > 0 {
		dto.Updater = fmt.Sprintf("%d", entity.Updater)
	}

	// 格式化时间字段
	if !entity.CreateDate.IsZero() {
		dto.CreateDate = entity.CreateDate.Format("2006-01-02 15:04:05")
	}
	if !entity.UpdateDate.IsZero() {
		dto.UpdateDate = entity.UpdateDate.Format("2006-01-02 15:04:05")
	}

	return dto
}

// readFirmwareFile 读取固件文件（支持绝对路径和相对路径查找）
func (uc *OtaUsecase) readFirmwareFile(firmwarePath, otaType, otaVersion string) ([]byte, string, error) {
	var filePath string
	var err error

	// 检查是否是绝对路径
	if filepath.IsAbs(firmwarePath) {
		filePath = firmwarePath
	} else {
		// 如果是相对路径，从当前工作目录解析
		wd, _ := os.Getwd()
		filePath = filepath.Join(wd, firmwarePath)
	}

	// 尝试读取文件
	fileData, err := os.ReadFile(filePath)
	if err == nil {
		// 文件读取成功，生成文件名
		filename := otaType + "_" + otaVersion
		if ext := filepath.Ext(firmwarePath); ext != "" {
			filename += ext
		}
		// 清理文件名，移除不安全字符
		filename = sanitizeFilename(filename)
		return fileData, filename, nil
	}

	// 如果文件不存在，尝试从firmware目录下查找文件名
	fileName := filepath.Base(firmwarePath)
	wd, _ := os.Getwd()
	altPath := filepath.Join(wd, "firmware", fileName)

	fileData, err = os.ReadFile(altPath)
	if err != nil {
		return nil, "", fmt.Errorf("固件文件不存在: %v", err)
	}

	// 文件读取成功，生成文件名
	filename := otaType + "_" + otaVersion
	if ext := filepath.Ext(fileName); ext != "" {
		filename += ext
	}
	// 清理文件名，移除不安全字符
	filename = sanitizeFilename(filename)
	return fileData, filename, nil
}

// sanitizeFilename 清理文件名，移除不安全字符
func sanitizeFilename(filename string) string {
	result := ""
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			result += string(r)
		} else {
			result += "_"
		}
	}
	return result
}

// calculateMD5 计算文件的MD5值（用于兼容旧代码，但当前使用内存中的文件数据）
func calculateMD5(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CheckDeviceActive 检查设备激活状态和版本
func (uc *OtaUsecase) CheckDeviceActive(ctx context.Context, macAddress, clientId string, deviceReport *DeviceReportReqDTO) (*DeviceReportRespDTO, error) {
	response := &DeviceReportRespDTO{}
	response.ServerTime = uc.BuildServerTime()

	// 查询设备是否存在
	device, err := uc.DeviceUsecase.GetDeviceByMacAddress(ctx, macAddress)
	if err != nil {
		return nil, err
	}

	// 设备未绑定，则返回当前上传的固件信息（不更新）以此兼容旧固件版本
	if device == nil {
		firmware := &Firmware{}
		if deviceReport.Application != nil && deviceReport.Application.Version != nil {
			firmware.Version = *deviceReport.Application.Version
		}
		firmware.URL = "http://xiaozhi.server.com:8002/xiaozhi/otaMag/download/NOT_ACTIVATED_FIRMWARE_THIS_IS_A_INVALID_URL"
		response.Firmware = firmware
	} else {
		// 只有在设备已绑定且autoUpdate不为0的情况下才返回固件升级信息
		if device.AutoUpdate != 0 {
			var deviceType string
			if deviceReport.Board != nil && deviceReport.Board.Type != nil {
				deviceType = *deviceReport.Board.Type
			}
			var currentVersion string
			if deviceReport.Application != nil && deviceReport.Application.Version != nil {
				currentVersion = *deviceReport.Application.Version
			}
			firmware := uc.BuildFirmwareInfo(ctx, deviceType, currentVersion)
			if firmware != nil {
				response.Firmware = firmware
			}
		}
	}

	// 添加WebSocket配置
	websocket := &Websocket{}
	wsUrl, err := uc.ConfigUsecase.GetValue(ctx, "server.websocket", true)
	if err != nil || wsUrl == "" || wsUrl == "null" {
		uc.log.Error("WebSocket地址未配置，请登录智控台，在参数管理找到【server.websocket】配置")
		wsUrl = "ws://xiaozhi.server.com:8000/xiaozhi/v1/"
		websocket.URL = wsUrl
	} else {
		wsUrls := strings.Split(wsUrl, ";")
		if len(wsUrls) > 0 {
			// 随机选择一个WebSocket URL
			randomIndex := rand.Intn(len(wsUrls))
			websocket.URL = wsUrls[randomIndex]
		} else {
			uc.log.Error("WebSocket地址未配置，请登录智控台，在参数管理找到【server.websocket】配置")
			websocket.URL = "ws://xiaozhi.server.com:8000/xiaozhi/v1/"
		}
	}
	websocket.Token = ""
	response.Websocket = websocket

	// 添加MQTT UDP配置
	mqttUdpConfig, err := uc.ConfigUsecase.GetValue(ctx, "server.mqtt_gateway", false)
	if err == nil && mqttUdpConfig != "" && mqttUdpConfig != "null" {
		var groupId string
		if device != nil && device.Board != "" {
			groupId = device.Board
		} else {
			groupId = "GID_default"
		}
		mqtt, err := uc.BuildMqttConfig(ctx, macAddress, groupId)
		if err != nil {
			uc.log.Error("生成MQTT配置失败: %v", err)
		} else if mqtt != nil {
			mqtt.Endpoint = mqttUdpConfig
			response.MQTT = mqtt
		}
	}

	if device != nil {
		// 如果设备存在，则异步更新上次连接时间和版本信息
		var appVersion string
		if deviceReport.Application != nil && deviceReport.Application.Version != nil {
			appVersion = *deviceReport.Application.Version
		}
		// 异步更新设备连接信息（这里简化处理，直接同步更新）
		go func() {
			now := time.Now()
			device.LastConnectedAt = &now
			if appVersion != "" {
				device.AppVersion = appVersion
			}
			device.UpdateDate = now
			_ = uc.DeviceUsecase.repo.Update(ctx, device)
		}()
	} else {
		// 如果设备不存在，则生成激活码
		activation := uc.BuildActivation(ctx, macAddress, deviceReport)
		response.Activation = activation
	}

	return response, nil
}

// BuildActivation 构建激活码
func (uc *OtaUsecase) BuildActivation(ctx context.Context, deviceId string, deviceReport *DeviceReportReqDTO) *Activation {
	activation := &Activation{}

	// 检查是否已有激活码
	cachedCode, _ := uc.DeviceUsecase.GeCodeByDeviceId(ctx, deviceId)

	if cachedCode != "" {
		activation.Code = cachedCode
		frontedUrl, _ := uc.ConfigUsecase.GetValue(ctx, "server.fronted_url", true)
		activation.Message = frontedUrl + "\n" + cachedCode
		activation.Challenge = deviceId
	} else {
		// 生成新的6位随机数字激活码
		newCode := fmt.Sprintf("%06d", rand.Intn(1000000))
		activation.Code = newCode
		frontedUrl, _ := uc.ConfigUsecase.GetValue(ctx, "server.fronted_url", true)
		activation.Message = frontedUrl + "\n" + newCode
		activation.Challenge = deviceId

		// 构建设备数据Map
		dataMap := make(map[string]interface{})
		dataMap["id"] = deviceId
		dataMap["mac_address"] = deviceId

		var board string
		if deviceReport.Board != nil && deviceReport.Board.Type != nil {
			board = *deviceReport.Board.Type
		} else if deviceReport.ChipModelName != nil {
			board = *deviceReport.ChipModelName
		} else {
			board = "unknown"
		}
		dataMap["board"] = board

		var appVersion string
		if deviceReport.Application != nil && deviceReport.Application.Version != nil {
			appVersion = *deviceReport.Application.Version
		}
		dataMap["app_version"] = appVersion
		dataMap["deviceId"] = deviceId
		dataMap["activation_code"] = newCode

		// 写入主数据 key
		safeDeviceId := strings.ToLower(strings.ReplaceAll(deviceId, ":", "_"))
		dataKey := fmt.Sprintf("ota:activation:data:%s", safeDeviceId)
		if err := uc.redisClient.SetObject(ctx, dataKey, dataMap, 0); err != nil {
			uc.log.Error("存储激活数据失败: %v", err)
		}

		// 写入反查激活码 key
		codeKey := fmt.Sprintf("ota:activation:code:%s", newCode)
		if err := uc.redisClient.Set(ctx, codeKey, deviceId, 0); err != nil {
			uc.log.Error("存储激活码映射失败: %v", err)
		}
	}

	return activation
}

// BuildFirmwareInfo 构建固件信息
func (uc *OtaUsecase) BuildFirmwareInfo(ctx context.Context, deviceType, currentVersion string) *Firmware {
	if deviceType == "" {
		return nil
	}
	if currentVersion == "" {
		currentVersion = "0.0.0"
	}

	ota, err := uc.repo.GetLatestOta(ctx, deviceType)
	if err != nil {
		uc.log.Error("查询最新固件失败: %v", err)
		return nil
	}

	firmware := &Firmware{}
	var downloadUrl string

	if ota != nil {
		// 如果设备没有版本信息，或者OTA版本比设备版本新，则返回下载地址
		if uc.CompareVersions(ota.Version, currentVersion) > 0 {
			otaUrl, err := uc.ConfigUsecase.GetValue(ctx, "server.ota", true)
			if err != nil || otaUrl == "" || otaUrl == "null" {
				uc.log.Error("OTA地址未配置，请登录智控台，在参数管理找到【server.ota】配置")
				// 这里无法从请求中获取URL，使用默认值
				otaUrl = "http://127.0.0.1:8001/manager-api/ota"
			}
			// 将URL中的/ota/替换为/otaMag/download/
			uuidStr := uuid.New().String()
			redisKey := kit.GetOtaIdKey(uuidStr)
			if err := uc.redisClient.Set(ctx, redisKey, ota.ID, 24*time.Hour); err != nil {
				uc.log.Error("存储OTA ID失败: %v", err)
			}
			downloadUrl = strings.Replace(otaUrl, "/ota/", "/otaMag/download/", 1) + uuidStr
		}
	}

	if ota == nil {
		firmware.Version = currentVersion
	} else {
		firmware.Version = ota.Version
	}
	if downloadUrl == "" {
		firmware.URL = "http://xiaozhi.server.com:8002/xiaozhi/otaMag/download/NOT_ACTIVATED_FIRMWARE_THIS_IS_A_INVALID_URL"
	} else {
		firmware.URL = downloadUrl
	}

	return firmware
}

// BuildMqttConfig 构建MQTT配置
func (uc *OtaUsecase) BuildMqttConfig(ctx context.Context, macAddress, groupId string) (*MQTT, error) {
	// 从系统参数获取签名密钥
	signatureKey, err := uc.ConfigUsecase.GetValue(ctx, "server.mqtt_signature_key", false)
	if err != nil || signatureKey == "" {
		uc.log.Warn("缺少MQTT_SIGNATURE_KEY，跳过MQTT配置生成")
		return nil, nil
	}

	// 构建客户端ID格式：groupId@@@macAddress@@@macAddress
	groupIdSafeStr := strings.ReplaceAll(groupId, ":", "_")
	deviceIdSafeStr := strings.ReplaceAll(macAddress, ":", "_")
	mqttClientId := fmt.Sprintf("%s@@@%s@@@%s", groupIdSafeStr, deviceIdSafeStr, deviceIdSafeStr)

	// 构建用户数据（包含IP等信息）
	userData := map[string]string{
		"ip": "unknown", // 这里无法从context获取IP，使用unknown
	}

	// 将用户数据编码为Base64 JSON
	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return nil, fmt.Errorf("编码用户数据失败: %w", err)
	}
	username := base64.StdEncoding.EncodeToString(userDataJSON)

	// 生成密码签名
	password, err := uc.generatePasswordSignature(mqttClientId+"|"+username, signatureKey)
	if err != nil {
		return nil, fmt.Errorf("生成密码签名失败: %w", err)
	}

	// 构建MQTT配置
	mqtt := &MQTT{
		ClientID:       mqttClientId,
		Username:       username,
		Password:       password,
		PublishTopic:   "device-server",
		SubscribeTopic: "devices/p2p/" + deviceIdSafeStr,
	}

	return mqtt, nil
}

// generatePasswordSignature 生成MQTT密码签名
func (uc *OtaUsecase) generatePasswordSignature(content, secretKey string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(content))
	signature := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// BuildServerTime 构建服务器时间
func (uc *OtaUsecase) BuildServerTime() *ServerTime {
	now := time.Now()
	tz := now.Location()
	serverTime := &ServerTime{
		Timestamp: now.UnixMilli(),
		TimeZone:  tz.String(),
	}

	// 计算时区偏移量（分钟）
	_, offset := now.Zone()
	serverTime.TimezoneOffset = int32(offset / 60)

	return serverTime
}

// CompareVersions 比较两个版本号
// 返回：1 (v1 > v2), -1 (v1 < v2), 0 (v1 == v2)
func (uc *OtaUsecase) CompareVersions(version1, version2 string) int {
	if version1 == "" || version2 == "" {
		return 0
	}

	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1, v2 int
		if i < len(v1Parts) {
			v1, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			v2, _ = strconv.Atoi(v2Parts[i])
		}

		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
	}

	return 0
}

// IsMacAddressValid 简单判断mac地址是否有效（非严格）
func IsMacAddressValid(macAddress string) bool {
	if macAddress == "" {
		return false
	}
	// MAC地址通常为12位十六进制数字，可以包含冒号或连字符分隔符
	pattern := `^([0-9A-Za-z]{2}[:-]){5}([0-9A-Za-z]{2})$`
	matched, _ := regexp.MatchString(pattern, macAddress)
	return matched
}

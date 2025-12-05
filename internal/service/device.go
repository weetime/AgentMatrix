package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

type DeviceService struct {
	uc            *biz.DeviceUsecase
	configUsecase *biz.ConfigUsecase
	redisClient   *redis.Client
	httpClient    *http.Client
	pb.UnimplementedDeviceServiceServer
}

func NewDeviceService(
	uc *biz.DeviceUsecase,
	configUsecase *biz.ConfigUsecase,
	redisClientWrapper *kit.RedisClient,
) *DeviceService {
	return &DeviceService{
		uc:            uc,
		configUsecase: configUsecase,
		redisClient:   redisClientWrapper.GetClient(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterDevice 设备注册（生成验证码）
func (s *DeviceService) RegisterDevice(ctx context.Context, req *pb.DeviceRegisterRequest) (*pb.Response, error) {
	macAddress := req.GetMacAddress()
	if macAddress == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "MAC地址不能为空",
		}, nil
	}

	// 生成6位验证码
	var code string
	var existsMac string
	for {
		code = kit.GenerateDeviceRegisterCode()
		key := kit.GetDeviceCaptchaKey(code)
		val, err := s.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			// 验证码不存在，可以使用
			break
		}
		if err != nil {
			return &pb.Response{
				Code: 500,
				Msg:  "生成验证码失败: " + err.Error(),
			}, nil
		}
		existsMac = val
		if existsMac == "" {
			// 验证码存在但值为空，可以使用
			break
		}
		// 验证码已存在，继续生成新的
	}

	// 存储到Redis
	key := kit.GetDeviceCaptchaKey(code)
	if err := s.redisClient.Set(ctx, key, macAddress, 0).Err(); err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "存储验证码失败: " + err.Error(),
		}, nil
	}

	data := map[string]interface{}{
		"code": code,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// BindDevice 绑定设备
func (s *DeviceService) BindDevice(ctx context.Context, req *pb.DeviceBindRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	agentId := req.GetAgentId()
	deviceCode := req.GetDeviceCode()

	// 从Redis获取激活码对应的设备ID
	deviceKey := kit.GetDeviceActivationCodeKey(deviceCode)
	deviceId, err := s.redisClient.Get(ctx, deviceKey).Result()
	if err == redis.Nil {
		return &pb.Response{
			Code: 400,
			Msg:  "激活码错误",
		}, nil
	}
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "获取激活码失败: " + err.Error(),
		}, nil
	}

	// 获取设备激活数据
	safeDeviceId := kit.SafeDeviceID(deviceId)
	cacheDeviceKey := kit.GetDeviceActivationDataKey(safeDeviceId)
	cacheDataStr, err := s.redisClient.Get(ctx, cacheDeviceKey).Result()
	if err == redis.Nil {
		return &pb.Response{
			Code: 400,
			Msg:  "激活码错误",
		}, nil
	}
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "获取激活数据失败: " + err.Error(),
		}, nil
	}

	// 解析缓存数据（Redis中存储的是JSON字符串）
	var cacheMap map[string]interface{}
	if err := json.Unmarshal([]byte(cacheDataStr), &cacheMap); err != nil {
		// 如果解析失败，尝试直接使用字符串值
		// 某些Redis客户端可能已经反序列化了
		return &pb.Response{
			Code: 500,
			Msg:  "解析激活数据失败: " + err.Error(),
		}, nil
	}

	// 验证激活码
	cachedCode, ok := cacheMap["activation_code"].(string)
	if !ok || cachedCode != deviceCode {
		return &pb.Response{
			Code: 400,
			Msg:  "激活码错误",
		}, nil
	}

	// 检查设备是否已被激活
	existingDevice, err := s.uc.GetDeviceByID(ctx, deviceId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "查询设备失败: " + err.Error(),
		}, nil
	}
	if existingDevice != nil {
		return &pb.Response{
			Code: 400,
			Msg:  "设备已被激活",
		}, nil
	}

	// 获取设备信息
	macAddress, _ := cacheMap["mac_address"].(string)
	board, _ := cacheMap["board"].(string)
	appVersion, _ := cacheMap["app_version"].(string)

	// 调用Usecase的DeviceActivation方法
	if err := s.uc.DeviceActivation(ctx, userID, agentId, deviceCode, deviceId, macAddress, board, appVersion); err != nil {
		return nil, err
	}

	// 清理Redis缓存
	s.redisClient.Del(ctx, cacheDeviceKey)
	s.redisClient.Del(ctx, deviceKey)
	s.redisClient.Del(ctx, kit.GetAgentDeviceCountKey(agentId))

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetUserDevices 获取已绑定设备
func (s *DeviceService) GetUserDevices(ctx context.Context, req *pb.GetUserDevicesRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	agentId := req.GetAgentId()
	devices, err := s.uc.GetUserDevices(ctx, userID, agentId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表，所有ID字段格式化为字符串
	deviceList := make([]interface{}, 0, len(devices))
	for _, device := range devices {
		deviceVO := map[string]interface{}{
			"id":         device.ID,
			"userId":     fmt.Sprintf("%d", device.UserID),
			"macAddress": device.MacAddress,
			"autoUpdate": device.AutoUpdate,
			"board":      device.Board,
			"alias":      device.Alias,
			"agentId":    device.AgentID,
			"appVersion": device.AppVersion,
			"sort":       device.Sort,
			"creator":    fmt.Sprintf("%d", device.Creator),
			"updater":    fmt.Sprintf("%d", device.Updater),
		}

		if device.LastConnectedAt != nil {
			deviceVO["lastConnectedAt"] = device.LastConnectedAt.Format("2006-01-02 15:04:05")
		}
		if !device.CreateDate.IsZero() {
			deviceVO["createDate"] = device.CreateDate.Format("2006-01-02 15:04:05")
		}
		if !device.UpdateDate.IsZero() {
			deviceVO["updateDate"] = device.UpdateDate.Format("2006-01-02 15:04:05")
		}

		deviceList = append(deviceList, deviceVO)
	}

	data := map[string]interface{}{
		"list": deviceList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// ForwardToMqttGateway 设备在线状态查询（转发MQTT）
func (s *DeviceService) ForwardToMqttGateway(ctx context.Context, req *pb.DeviceStatusRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	agentId := req.GetAgentId()

	// 从系统参数获取MQTT网关地址
	mqttGatewayUrl, err := s.configUsecase.GetValue(ctx, "server.mqtt_manager_api", true)
	if err != nil || mqttGatewayUrl == "" || mqttGatewayUrl == "null" {
		return &pb.Response{
			Code: 0,
			Msg:  "success",
		}, nil
	}

	// 获取当前用户的设备列表
	devices, err := s.uc.GetUserDevices(ctx, userID, agentId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建deviceIds数组
	deviceIds := make([]string, 0, len(devices))
	for _, device := range devices {
		macAddress := device.MacAddress
		if macAddress == "" {
			macAddress = "unknown"
		}
		groupId := device.Board
		if groupId == "" {
			groupId = "GID_default"
		}

		// 替换冒号为下划线
		groupId = strings.ReplaceAll(groupId, ":", "_")
		macAddress = strings.ReplaceAll(macAddress, ":", "_")

		// 构建mqtt客户端ID格式：groupId@@@macAddress@@@macAddress
		mqttClientId := fmt.Sprintf("%s@@@%s@@@%s", groupId, macAddress, macAddress)
		deviceIds = append(deviceIds, mqttClientId)
	}

	// 构建完整的URL
	url := fmt.Sprintf("http://%s/api/devices/status", mqttGatewayUrl)

	// 生成Bearer Token
	signatureKey, err := s.configUsecase.GetValue(ctx, "server.mqtt_signature_key", false)
	if err != nil || signatureKey == "" {
		return &pb.Response{
			Code: 500,
			Msg:  "令牌生成失败",
		}, nil
	}
	token := kit.GenerateBearerTokenForToday(signatureKey)

	// 构建请求体JSON
	requestBody := map[string]interface{}{
		"clientIds": deviceIds,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建请求体失败: " + err.Error(),
		}, nil
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "创建请求失败: " + err.Error(),
		}, nil
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "转发请求失败: " + err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "读取响应失败: " + err.Error(),
		}, nil
	}

	// 直接返回MQTT网关的响应体（与Java版本一致）
	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"result": {Kind: &structpb.Value_StringValue{StringValue: responseBody.String()}},
			},
		},
	}, nil
}

// UnbindDevice 解绑设备
func (s *DeviceService) UnbindDevice(ctx context.Context, req *pb.DeviceUnbindRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	deviceId := req.GetDeviceId()
	if err := s.uc.UnbindDevice(ctx, userID, deviceId); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// UpdateDeviceInfo 更新设备信息
func (s *DeviceService) UpdateDeviceInfo(ctx context.Context, req *pb.DeviceUpdateInfoRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	deviceId := req.GetId()
	var autoUpdate *int32
	var alias *string

	if req.AutoUpdate != nil {
		val := req.AutoUpdate.GetValue()
		autoUpdate = &val
	}
	if req.Alias != nil {
		val := req.Alias.GetValue()
		alias = &val
	}

	if err := s.uc.UpdateDevice(ctx, deviceId, userID, autoUpdate, alias); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// ManualAddDevice 手动添加设备
func (s *DeviceService) ManualAddDevice(ctx context.Context, req *pb.DeviceManualAddRequest) (*pb.Response, error) {
	// 从context获取当前用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "user not authenticated",
		}, nil
	}

	if err := s.uc.ManualAddDevice(ctx, userID, req.GetAgentId(), req.GetBoard(), req.GetAppVersion(), req.GetMacAddress()); err != nil {
		return nil, err
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

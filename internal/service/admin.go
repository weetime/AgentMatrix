package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer
	userUsecase   *biz.UserUsecase
	configUsecase *biz.ConfigUsecase
}

func NewAdminService(
	userUsecase *biz.UserUsecase,
	configUsecase *biz.ConfigUsecase,
) *AdminService {
	return &AdminService{
		userUsecase:   userUsecase,
		configUsecase: configUsecase,
	}
}

// PageAdminUsers 分页查找用户
func (s *AdminService) PageAdminUsers(ctx context.Context, req *pb.PageAdminUsersRequest) (*pb.Response, error) {
	// 解析过滤条件
	var mobile *string
	if req.Mobile != nil && req.Mobile.GetValue() != "" {
		m := req.Mobile.GetValue()
		mobile = &m
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("id")

	// 查询列表
	list, total, err := s.userUsecase.PageAdminUsers(ctx, mobile, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"userid":      item.UserID,
			"mobile":      item.Mobile,
			"deviceCount": item.DeviceCount,
			"status":      item.Status,
			"createDate":  item.CreateDate.Format("2006-01-02 15:04:05"),
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
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

// ResetUserPassword 重置用户密码
func (s *AdminService) ResetUserPassword(ctx context.Context, req *pb.ResetUserPasswordRequest) (*pb.Response, error) {
	userId, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &pb.Response{
			Code: 400,
			Msg:  "无效的用户ID",
		}, nil
	}

	password, err := s.userUsecase.ResetPassword(ctx, userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 返回密码字符串（直接作为data字段的值，与Java实现一致）
	// Java: Result<String>.ok(password) -> data字段直接是字符串
	// 前端代码：data.data 直接读取密码字符串
	// 在protobuf中，Response的Data是Struct类型，我们需要将字符串包装成Struct
	// 根据前端的使用方式，data.data应该直接是字符串值
	// 我们使用一个简单的key来存储密码，前端会直接读取data.data的值
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"": password, // 使用空字符串作为key，前端会直接读取data.data
	})
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

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.Response, error) {
	userId, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &pb.Response{
			Code: 400,
			Msg:  "无效的用户ID",
		}, nil
	}

	err = s.userUsecase.DeleteUserById(ctx, userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// ChangeUserStatus 批量修改用户状态
func (s *AdminService) ChangeUserStatus(ctx context.Context, req *pb.ChangeUserStatusRequest) (*pb.Response, error) {
	status := req.GetStatus()
	userIds := req.GetUserIds()

	if len(userIds) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "用户ID列表不能为空",
		}, nil
	}

	err := s.userUsecase.ChangeUserStatus(ctx, status, userIds)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetServerList 获取WebSocket服务端列表
func (s *AdminService) GetServerList(ctx context.Context, req *pb.GetServerListRequest) (*pb.Response, error) {
	// 获取 server.websocket 配置
	wsText, err := s.configUsecase.GetValue(ctx, "server.websocket", true)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "获取配置失败: " + err.Error(),
		}, nil
	}

	// 如果为空，返回空数组
	var serverList []string
	if wsText != "" {
		// 按 ; 分割字符串
		splitList := strings.Split(wsText, ";")
		// 过滤空字符串
		serverList = make([]string, 0, len(splitList))
		for _, server := range splitList {
			if trimmed := strings.TrimSpace(server); trimmed != "" {
				serverList = append(serverList, trimmed)
			}
		}
	}

	// 将 []string 转换为 []interface{}，因为 structpb.NewStruct 不接受 []string
	serverListInterface := make([]interface{}, len(serverList))
	for i, server := range serverList {
		serverListInterface[i] = server
	}

	// 返回字符串列表（与Java的Result<List<String>>一致）
	// 使用 "list" 作为key，将数组作为值，这样前端可以通过 data.data.list 访问数组
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"list":  serverListInterface,
		"total": int32(len(serverListInterface)),
	})
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

// EmitServerAction 通知服务端更新配置
func (s *AdminService) EmitServerAction(ctx context.Context, req *pb.EmitServerActionRequest) (*pb.Response, error) {
	// 验证 action 不为空
	action := req.GetAction()
	if action == "" {
		return &pb.Response{
			Code: int32(kit.ErrInvalidServerAction),
			Msg:  "操作不能为空",
		}, nil
	}

	// 验证 action 值
	if action != "restart" && action != "update_config" {
		return &pb.Response{
			Code: int32(kit.ErrInvalidServerAction),
			Msg:  "无效的服务端操作",
		}, nil
	}

	// 获取 server.websocket 配置
	wsText, err := s.configUsecase.GetValue(ctx, "server.websocket", true)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "获取配置失败: " + err.Error(),
		}, nil
	}

	if wsText == "" {
		return &pb.Response{
			Code: int32(kit.ErrServerWebsocketNotConfigured),
			Msg:  "未配置服务端WebSocket地址",
		}, nil
	}

	// 验证 targetWs 在列表中
	targetWs := req.GetTargetWs()
	wsList := strings.Split(wsText, ";")
	found := false
	for _, ws := range wsList {
		if strings.TrimSpace(ws) == targetWs {
			found = true
			break
		}
	}

	if !found {
		return &pb.Response{
			Code: int32(kit.ErrTargetWebsocketNotExist),
			Msg:  "目标WebSocket地址不存在",
		}, nil
	}

	// 调用 WebSocket 发送动作指令
	success, err := s.emitServerActionByWs(ctx, targetWs, action)
	if err != nil {
		return &pb.Response{
			Code: int32(kit.ErrWebSocketConnectFailed),
			Msg:  "WebSocket连接失败: " + err.Error(),
		}, nil
	}

	// 返回布尔值（与Java的Result<Boolean>一致）
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"": success,
	})
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

// emitServerActionByWs 通过WebSocket发送服务端动作
func (s *AdminService) emitServerActionByWs(ctx context.Context, targetWsUri, action string) (bool, error) {
	if targetWsUri == "" || action == "" {
		return false, fmt.Errorf("参数不能为空")
	}

	// 获取 server.secret 配置
	serverSK, err := s.configUsecase.GetValue(ctx, "server.secret", true)
	if err != nil {
		return false, fmt.Errorf("获取服务器密钥失败: %w", err)
	}

	// 创建WebSocket客户端配置
	config := &kit.WebSocketClientConfig{
		URL:            targetWsUri,
		ConnectTimeout: 3 * time.Second,
		MaxDuration:    120 * time.Second,
		BufferSize:     8 * 1024,
		Logger:         nil, // 使用默认logger
	}

	client := kit.NewWebSocketClient(config)

	// 设置请求头
	headers := make(http.Header)
	headers.Set("device-id", uuid.New().String())
	headers.Set("client-id", uuid.New().String())

	// 连接WebSocket
	if err := client.ConnectWithHeaders(headers); err != nil {
		return false, fmt.Errorf("连接失败: %w", err)
	}

	// 确保关闭连接
	defer client.Close()

	// 构建并发送JSON消息
	payload := map[string]interface{}{
		"type":   "server",
		"action": action,
		"content": map[string]interface{}{
			"secret": serverSK,
		},
	}

	if err := client.SendJson(payload); err != nil {
		return false, fmt.Errorf("发送消息失败: %w", err)
	}

	// 监听响应
	responses, err := client.ListenForJsonResponse(func(resp map[string]interface{}) bool {
		// 验证响应：status == "success" 且 type == "server" 且 content.action 存在
		status, ok1 := resp["status"].(string)
		respType, ok2 := resp["type"].(string)
		content, ok3 := resp["content"].(map[string]interface{})

		if !ok1 || !ok2 || !ok3 {
			return false
		}

		if status != "success" {
			return false
		}

		if respType != "server" {
			return false
		}

		_, hasAction := content["action"]
		return hasAction
	})

	if err != nil {
		return false, fmt.Errorf("监听响应失败: %w", err)
	}

	// 如果收到匹配的响应，返回true
	return len(responses) > 0, nil
}

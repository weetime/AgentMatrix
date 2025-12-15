package service

import (
	"context"
	"strings"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type ConfigService struct {
	uc *biz.ConfigUsecase
	pb.UnimplementedConfigServiceServer
}

func NewConfigService(uc *biz.ConfigUsecase) *ConfigService {
	return &ConfigService{
		uc: uc,
	}
}

// GetConfigUsecase 获取 ConfigUsecase（用于创建 ServerSecretServiceAdapter）
func (s *ConfigService) GetConfigUsecase() *biz.ConfigUsecase {
	return s.uc
}

// GetServerConfig 获取服务器配置
func (s *ConfigService) GetServerConfig(ctx context.Context, req *pb.GetServerConfigRequest) (*pb.Response, error) {
	// 从请求中读取 is_cache 参数，如果未设置则默认为 true（与 Java 实现保持一致）
	isCache := true
	if req.GetIsCache() != nil {
		isCache = req.GetIsCache().GetValue()
	}

	// 获取配置
	config, err := s.uc.GetConfig(ctx, isCache)
	if err != nil {
		return nil, err
	}

	// 转换为 protobuf Struct
	configStruct, err := structpb.NewStruct(config)
	if err != nil {
		return nil, err
	}

	// 构建统一响应格式
	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: configStruct,
	}, nil
}

// ServerSecretServiceAdapter 将 ConfigUsecase 适配为 ServerSecretService
type ServerSecretServiceAdapter struct {
	uc *biz.ConfigUsecase
}

// NewServerSecretServiceAdapter 创建 ServerSecretService 适配器
func NewServerSecretServiceAdapter(uc *biz.ConfigUsecase) middleware.ServerSecretService {
	return &ServerSecretServiceAdapter{uc: uc}
}

// GetServerSecret 获取服务器密钥（实现 ServerSecretService 接口）
func (a *ServerSecretServiceAdapter) GetServerSecret(ctx context.Context) (string, error) {
	return a.uc.GetValue(ctx, "server.secret", true)
}

// GetAgentModels 获取智能体模型配置
func (s *ConfigService) GetAgentModels(ctx context.Context, req *pb.GetAgentModelsRequest) (*pb.Response, error) {
	// 验证请求参数
	if req.GetMacAddress() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "设备MAC地址不能为空",
		}, nil
	}
	if req.GetClientId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "客户端ID不能为空",
		}, nil
	}
	if req.GetSelectedModule() == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "客户端已实例化的模型不能为空",
		}, nil
	}

	// 转换 selectedModule map
	selectedModule := make(map[string]string)
	for k, v := range req.GetSelectedModule() {
		selectedModule[k] = v
	}

	// 调用业务逻辑层
	result, err := s.uc.GetAgentModels(ctx, req.GetMacAddress(), selectedModule)
	if err != nil {
		// 检查是否是特定的错误码
		errMsg := err.Error()
		if strings.Contains(errMsg, "设备需要绑定") {
			// 提取激活码
			parts := strings.Split(errMsg, "激活码: ")
			if len(parts) > 1 {
				return &pb.Response{
					Code: 10042, // OTA_DEVICE_NEED_BIND
					Msg:  errMsg,
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"activation_code": structpb.NewStringValue(parts[1]),
						},
					},
				}, nil
			}
		} else if strings.Contains(errMsg, "设备未找到") {
			return &pb.Response{
				Code: 10041, // OTA_DEVICE_NOT_FOUND
				Msg:  errMsg,
			}, nil
		} else if strings.Contains(errMsg, "智能体未找到") {
			return &pb.Response{
				Code: 10053, // AGENT_NOT_FOUND
				Msg:  errMsg,
			}, nil
		}
		return &pb.Response{
			Code: 500,
			Msg:  errMsg,
		}, nil
	}

	// 转换为 protobuf Struct
	configStruct, err := structpb.NewStruct(result)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	// 构建统一响应格式
	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: configStruct,
	}, nil
}

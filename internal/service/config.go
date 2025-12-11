package service

import (
	"context"

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

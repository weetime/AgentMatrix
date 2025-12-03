package service

import (
	"context"

	"nova/internal/biz"
	pb "nova/protos/nova/v1"

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

// GetServerConfig 获取服务器配置
func (s *ConfigService) GetServerConfig(ctx context.Context, req *pb.GetServerConfigRequest) (*pb.GetServerConfigResponse, error) {
	// 默认从缓存读取
	isCache := true
	if req != nil {
		isCache = req.IsCache
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

	return &pb.GetServerConfigResponse{
		Config: configStruct,
	}, nil
}


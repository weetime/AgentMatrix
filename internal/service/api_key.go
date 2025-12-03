package service

import (
	"context"

	"nova/internal/biz"
	"nova/internal/kit"
	pb "nova/protos/nova/v1"

	"github.com/jinzhu/copier"
)

type ApiKeyService struct {
	uc *biz.ApiKeyUsecase
	pb.UnimplementedApiKeyServiceServer
}

func NewApiKeyService(uc *biz.ApiKeyUsecase) *ApiKeyService {
	return &ApiKeyService{
		uc: uc,
	}
}

func (s *ApiKeyService) Create(ctx context.Context, req *pb.WorkspaceName) (*pb.ApiKey, error) {
	apiKey, err := s.uc.Create(ctx, "admin", req.GetWorkspaceName())
	if err != nil {
		return nil, err
	}
	key := &pb.ApiKey{}
	if err := copier.Copy(key, apiKey); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *ApiKeyService) List(ctx context.Context, req *pb.ListApiKeyReq) (*pb.ApiKeys, error) {
	filters := req.GetFilters()
	param := &biz.ListApiKeyParams{}
	if filters != nil {
		if err := copier.Copy(param, filters); err != nil {
			return nil, err
		}
	}

	page := &kit.PageRequest{}
	apiToPageRequest(page, req.GetPageRequest())

	list, err := s.uc.List(ctx, param, page)
	if err != nil {
		return nil, err
	}

	result := &pb.ApiKeys{List: []*pb.ApiKey{}}
	if err := copier.Copy(&result.List, list); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ApiKeyService) TotalCount(ctx context.Context, req *pb.ListApiKeyReq) (*pb.Total, error) {
	param := &biz.ListApiKeyParams{}
	if err := copier.Copy(param, req); err != nil {
		return nil, err
	}

	total, err := s.uc.TotalCount(ctx, param)
	if err != nil {
		return nil, err
	}

	return &pb.Total{Total: int64(total)}, nil
}

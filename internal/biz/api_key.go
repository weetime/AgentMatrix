package biz

import (
	"context"
	"errors"
	"fmt"

	"nova/internal/kit"
	"nova/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
)

type ApiKey struct {
	ID            int64
	UUID          uuid.UUID
	Key           string `validate:"required"`
	Username      string `validate:"required"`
	WorkspaceName string `validate:"required"`
	Name          string `validate:"required"`
	Models        string
	IsEnabled     bool
	IsDeleted     bool
}

type ListApiKeyParams struct {
	Username      *wrappers.StringValue
	WorkspaceName *wrappers.StringValue
	Name          *wrappers.StringValue
	Model         *wrappers.StringValue
	IsEnabled     *wrappers.BoolValue
}

type ApiKeyRepo interface {
	Create(ctx context.Context, apiKey *ApiKey) error
	Exist(ctx context.Context, uuid uuid.UUID) (bool, error)
	Total(ctx context.Context, params *ListApiKeyParams) (int, error)
	List(ctx context.Context, params *ListApiKeyParams, page *kit.PageRequest) ([]*ApiKey, error)
}

type ApiKeyUsecase struct {
	repo        ApiKeyRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

func NewApiKeyUsecase(
	repo ApiKeyRepo,
	logger log.Logger,
) *ApiKeyUsecase {
	return &ApiKeyUsecase{
		repo:        repo,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

func (uc *ApiKeyUsecase) generateApiKey() string {
	return fmt.Sprintf("sk-%s", uuid.New().String())
}

func (uc *ApiKeyUsecase) Create(ctx context.Context, username, workspaceName string) (*ApiKey, error) {
	if exist, err := uc.repo.Exist(ctx, kit.GeneratorUUID(username, workspaceName)); err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	} else if exist {
		return nil, uc.handleError.ErrAlreadyExists(ctx, errors.New("api_key already exists"))
	}

	apiKey := &ApiKey{
		UUID:          kit.GeneratorUUID(username, workspaceName),
		Key:           uc.generateApiKey(),
		Name:          username + "/" + workspaceName,
		Username:      username,
		WorkspaceName: workspaceName,
		Models:        "",
		IsEnabled:     true,
		IsDeleted:     false,
	}
	if err := kit.Validate(apiKey); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}

	if err := uc.repo.Create(ctx, apiKey); err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	return apiKey, nil
}

func (uc *ApiKeyUsecase) List(ctx context.Context, params *ListApiKeyParams, page *kit.PageRequest) ([]*ApiKey, error) {
	if err := kit.Validate(params); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.List(ctx, params, page)
}

func (uc *ApiKeyUsecase) TotalCount(ctx context.Context, params *ListApiKeyParams) (int, error) {
	if err := kit.Validate(params); err != nil {
		return 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.Total(ctx, params)
}

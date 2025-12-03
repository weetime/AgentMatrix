package data

import (
	"context"

	"nova/internal/biz"
	"nova/internal/data/ent"
	"nova/internal/data/ent/apikey"
	"nova/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type apiKeyRepo struct {
	data *Data
	log  *log.Helper
}

// NewApiKeyRepo 初始化Repo
func NewApiKeyRepo(data *Data, logger log.Logger) biz.ApiKeyRepo {
	return &apiKeyRepo{
		data: data,
		log:  kit.LogHelper(logger),
	}
}

func (r *apiKeyRepo) Exist(ctx context.Context, uuid uuid.UUID) (bool, error) {
	return r.data.db.ApiKey.Query().Where(apikey.UUID(uuid)).Exist(ctx)
}

func (r *apiKeyRepo) Create(ctx context.Context, bizApiKey *biz.ApiKey) error {
	return r.data.db.ApiKey.Create().
		SetUUID(bizApiKey.UUID).
		SetKey(bizApiKey.Key).
		SetUsername(bizApiKey.Username).
		SetWorkspaceName(bizApiKey.WorkspaceName).
		SetName(bizApiKey.Name).
		SetModels(bizApiKey.Models).
		SetIsEnabled(bizApiKey.IsEnabled).
		SetIsDeleted(bizApiKey.IsDeleted).
		Exec(ctx)
}

func (r *apiKeyRepo) Update(ctx context.Context, bizApiKey *biz.ApiKey) error {
	return r.data.db.ApiKey.Update().
		Where(apikey.UUID(bizApiKey.UUID)).
		SetName(bizApiKey.Name).
		SetModels(bizApiKey.Models).
		SetIsEnabled(bizApiKey.IsEnabled).
		Exec(ctx)
}

func (r *apiKeyRepo) Detail(ctx context.Context, uuid uuid.UUID) (*biz.ApiKey, error) {
	res, err := r.data.db.ApiKey.Query().Where(apikey.UUID(uuid)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return kit.AutoCopy(new(biz.ApiKey), res)
}

func (r *apiKeyRepo) Total(ctx context.Context, params *biz.ListApiKeyParams) (int, error) {
	query := r.data.db.ApiKey.Query()
	return r.applyFilters(query, params).Count(ctx)
}

func (r *apiKeyRepo) List(ctx context.Context, params *biz.ListApiKeyParams, page *kit.PageRequest) ([]*biz.ApiKey, error) {
	list, err := r.buildQuery(params, page).All(ctx)
	if err != nil {
		return nil, err
	}

	apiKeys := make([]*biz.ApiKey, len(list))
	if err := copier.Copy(&apiKeys, list); err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (r *apiKeyRepo) buildQuery(params *biz.ListApiKeyParams, page *kit.PageRequest) *ent.ApiKeyQuery {
	query := r.data.db.ApiKey.Query()
	query = r.applyFilters(query, params)

	applyPagination(query, page, apikey.Columns)

	return query
}

func (r *apiKeyRepo) applyFilters(query *ent.ApiKeyQuery, params *biz.ListApiKeyParams) *ent.ApiKeyQuery {
	if params.Username != nil {
		query.Where(apikey.Username(params.Username.GetValue()))
	}
	if params.Name != nil {
		query.Where(apikey.NameContains(params.Name.GetValue()))
	}
	if params.WorkspaceName != nil {
		query.Where(apikey.WorkspaceName(params.WorkspaceName.GetValue()))
	}
	if params.Model != nil {
		query.Where(apikey.ModelsContains(params.Model.GetValue()))
	}
	if params.IsEnabled != nil {
		query.Where(apikey.IsEnabled(params.IsEnabled.GetValue()))
	}

	return query.Where(apikey.IsDeleted(false))
}

package data

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/agent"
	"github.com/weetime/agent-matrix/internal/data/ent/modelconfig"
	"github.com/weetime/agent-matrix/internal/data/ent/modelprovider"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type modelConfigRepo struct {
	data *Data
	log  *log.Helper
}

type modelProviderRepo struct {
	data *Data
	log  *log.Helper
}

// NewModelConfigRepo 初始化模型配置Repo
func NewModelConfigRepo(data *Data, logger log.Logger) biz.ModelConfigRepo {
	return &modelConfigRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/model-config")),
	}
}

// NewModelProviderRepo 初始化模型供应器Repo
func NewModelProviderRepo(data *Data, logger log.Logger) biz.ModelProviderRepo {
	return &modelProviderRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/model-provider")),
	}
}

// GetModelCodeList 获取模型名称列表（仅查询启用的）
func (r *modelConfigRepo) GetModelCodeList(ctx context.Context, modelType string, modelName *string) ([]*biz.ModelConfig, error) {
	query := r.data.db.ModelConfig.Query().
		Where(
			modelconfig.ModelTypeEQ(modelType),
			modelconfig.IsEnabledEQ(true),
		)

	if modelName != nil && *modelName != "" {
		query = query.Where(modelconfig.ModelNameContains(*modelName))
	}

	// 只查询id和model_name字段
	list, err := query.Select(modelconfig.FieldID, modelconfig.FieldModelName).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelConfig, len(list))
	for i, item := range list {
		result[i] = &biz.ModelConfig{
			ID:        item.ID,
			ModelName: item.ModelName,
		}
	}

	return result, nil
}

// GetLlmModelCodeList 获取LLM模型信息列表
func (r *modelConfigRepo) GetLlmModelCodeList(ctx context.Context, modelName *string) ([]*biz.ModelConfig, error) {
	query := r.data.db.ModelConfig.Query().
		Where(
			modelconfig.ModelTypeEQ("LLM"),
			modelconfig.IsEnabledEQ(true),
		)

	if modelName != nil && *modelName != "" {
		query = query.Where(modelconfig.ModelNameContains(*modelName))
	}

	// 查询id、model_name和config_json字段
	list, err := query.Select(modelconfig.FieldID, modelconfig.FieldModelName, modelconfig.FieldConfigJSON).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelConfig, len(list))
	for i, item := range list {
		result[i] = &biz.ModelConfig{
			ID:         item.ID,
			ModelName:  item.ModelName,
			ConfigJSON: item.ConfigJSON,
		}
	}

	return result, nil
}

// GetModelConfigList 分页查询模型配置列表
func (r *modelConfigRepo) GetModelConfigList(ctx context.Context, params *biz.ListModelConfigParams, page *kit.PageRequest) ([]*biz.ModelConfig, error) {
	query := r.buildModelConfigQuery(params, page)
	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelConfig, len(list))
	for i, item := range list {
		result[i] = r.entityToBiz(item)
	}

	return result, nil
}

// TotalModelConfigs 获取模型配置总数
func (r *modelConfigRepo) TotalModelConfigs(ctx context.Context, params *biz.ListModelConfigParams) (int, error) {
	query := r.data.db.ModelConfig.Query()
	return r.applyModelConfigFilters(query, params).Count(ctx)
}

// buildModelConfigQuery 构建查询
func (r *modelConfigRepo) buildModelConfigQuery(params *biz.ListModelConfigParams, page *kit.PageRequest) *ent.ModelConfigQuery {
	query := r.data.db.ModelConfig.Query()
	query = r.applyModelConfigFilters(query, params)

	// 排序规则：先按is_enabled降序，再按sort升序
	query = query.Order(ent.Desc(modelconfig.FieldIsEnabled), ent.Asc(modelconfig.FieldSort))

	// 应用分页
	if page != nil {
		pageNo, _ := page.GetPageNo()
		pageSize := page.GetPageSize()
		if pageSize > 0 {
			query = query.Limit(pageSize)
			if pageNo > 0 {
				query = query.Offset((pageNo - 1) * pageSize)
			}
		}
	}

	return query
}

// applyModelConfigFilters 应用过滤条件
func (r *modelConfigRepo) applyModelConfigFilters(query *ent.ModelConfigQuery, params *biz.ListModelConfigParams) *ent.ModelConfigQuery {
	if params == nil {
		return query
	}

	if params.ModelType != "" {
		query = query.Where(modelconfig.ModelTypeEQ(params.ModelType))
	}

	if params.ModelName != nil && *params.ModelName != "" {
		query = query.Where(modelconfig.ModelNameContains(*params.ModelName))
	}

	return query
}

// GetModelConfigByID 获取模型配置（经过敏感数据处理）
func (r *modelConfigRepo) GetModelConfigByID(ctx context.Context, id string) (*biz.ModelConfig, error) {
	config, err := r.data.db.ModelConfig.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	bizConfig := r.entityToBiz(config)

	// 对configJson进行敏感数据掩码处理
	if bizConfig.ConfigJSON != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(bizConfig.ConfigJSON), &configMap); err == nil {
			maskedMap, err := kit.MaskSensitiveFieldsInMap(configMap)
			if err == nil {
				maskedJSON, err := json.Marshal(maskedMap)
				if err == nil {
					bizConfig.ConfigJSON = string(maskedJSON)
				}
			}
		}
	}

	return bizConfig, nil
}

// GetModelConfigByIDRaw 获取模型配置（不经过敏感数据处理）
func (r *modelConfigRepo) GetModelConfigByIDRaw(ctx context.Context, id string) (*biz.ModelConfig, error) {
	config, err := r.data.db.ModelConfig.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.entityToBiz(config), nil
}

// CreateModelConfig 创建模型配置
func (r *modelConfigRepo) CreateModelConfig(ctx context.Context, config *biz.ModelConfig) (*biz.ModelConfig, error) {
	create := r.data.db.ModelConfig.Create().
		SetID(config.ID).
		SetModelType(config.ModelType).
		SetModelCode(config.ModelCode).
		SetModelName(config.ModelName).
		SetIsDefault(config.IsDefault).
		SetIsEnabled(config.IsEnabled).
		SetSort(config.Sort).
		SetCreateDate(config.CreateDate).
		SetUpdateDate(config.UpdateDate)

	if config.ConfigJSON != "" {
		create.SetConfigJSON(config.ConfigJSON)
	}
	if config.DocLink != "" {
		create.SetDocLink(config.DocLink)
	}
	if config.Remark != "" {
		create.SetRemark(config.Remark)
	}
	if config.Creator > 0 {
		create.SetCreator(config.Creator)
	}
	if config.Updater > 0 {
		create.SetUpdater(config.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.entityToBiz(entity), nil
}

// UpdateModelConfig 更新模型配置
func (r *modelConfigRepo) UpdateModelConfig(ctx context.Context, config *biz.ModelConfig) error {
	update := r.data.db.ModelConfig.UpdateOneID(config.ID).
		SetModelType(config.ModelType).
		SetModelCode(config.ModelCode).
		SetModelName(config.ModelName).
		SetIsDefault(config.IsDefault).
		SetIsEnabled(config.IsEnabled).
		SetSort(config.Sort).
		SetUpdateDate(config.UpdateDate)

	// ConfigJSON的处理：如果为空字符串，不更新该字段（保持原值）
	// 注意：在EnableModelConfig和SetDefaultModel中，会传入空字符串表示不更新
	if config.ConfigJSON != "" {
		update.SetConfigJSON(config.ConfigJSON)
	}
	if config.DocLink != "" {
		update.SetDocLink(config.DocLink)
	} else {
		update.ClearDocLink()
	}
	if config.Remark != "" {
		update.SetRemark(config.Remark)
	} else {
		update.ClearRemark()
	}
	if config.Updater > 0 {
		update.SetUpdater(config.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteModelConfig 删除模型配置
func (r *modelConfigRepo) DeleteModelConfig(ctx context.Context, id string) error {
	return r.data.db.ModelConfig.DeleteOneID(id).Exec(ctx)
}

// SetDefaultModel 设置默认模型
func (r *modelConfigRepo) SetDefaultModel(ctx context.Context, modelType string, isDefault bool) error {
	_, err := r.data.db.ModelConfig.Update().
		Where(modelconfig.ModelTypeEQ(modelType)).
		SetIsDefault(isDefault).
		Save(ctx)
	return err
}

// CheckAgentReference 检查智能体引用
func (r *modelConfigRepo) CheckAgentReference(ctx context.Context, modelId string) ([]string, error) {
	agents, err := r.data.db.Agent.Query().
		Where(
			agent.Or(
				agent.AsrModelIDEQ(modelId),
				agent.VadModelIDEQ(modelId),
				agent.LlmModelIDEQ(modelId),
				agent.TtsModelIDEQ(modelId),
				agent.MemModelIDEQ(modelId),
				agent.VllmModelIDEQ(modelId),
				agent.IntentModelIDEQ(modelId),
			),
		).
		Select(agent.FieldAgentName).
		All(ctx)
	if err != nil {
		return nil, err
	}

	agentNames := make([]string, len(agents))
	for i, a := range agents {
		agentNames[i] = a.AgentName
	}

	return agentNames, nil
}

// CheckIntentConfigReference 检查意图识别配置引用
func (r *modelConfigRepo) CheckIntentConfigReference(ctx context.Context, modelId string) (bool, error) {
	// 查询Intent类型的模型配置，检查config_json中是否包含该modelId
	configs, err := r.data.db.ModelConfig.Query().
		Where(
			modelconfig.ModelTypeEQ("Intent"),
		).
		Select(modelconfig.FieldConfigJSON).
		All(ctx)
	if err != nil {
		return false, err
	}

	for _, config := range configs {
		if config.ConfigJSON != "" && strings.Contains(config.ConfigJSON, modelId) {
			return true, nil
		}
	}

	return false, nil
}

// entityToBiz 实体转换为业务对象
func (r *modelConfigRepo) entityToBiz(entity *ent.ModelConfig) *biz.ModelConfig {
	return &biz.ModelConfig{
		ID:         entity.ID,
		ModelType:  entity.ModelType,
		ModelCode:  entity.ModelCode,
		ModelName:  entity.ModelName,
		IsDefault:  entity.IsDefault,
		IsEnabled:  entity.IsEnabled,
		ConfigJSON: entity.ConfigJSON,
		DocLink:    entity.DocLink,
		Remark:     entity.Remark,
		Sort:       entity.Sort,
		Creator:    entity.Creator,
		CreateDate: entity.CreateDate,
		Updater:    entity.Updater,
		UpdateDate: entity.UpdateDate,
	}
}

// GetListByModelType 根据模型类型获取供应器列表
func (r *modelProviderRepo) GetListByModelType(ctx context.Context, modelType string) ([]*biz.ModelProvider, error) {
	list, err := r.data.db.ModelProvider.Query().
		Where(modelprovider.ModelTypeEQ(modelType)).
		Order(ent.Asc(modelprovider.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelProvider, len(list))
	for i, item := range list {
		result[i] = r.entityToBizModelProvider(item)
	}

	return result, nil
}

// GetList 根据模型类型和供应器代码获取供应器列表
func (r *modelProviderRepo) GetList(ctx context.Context, modelType string, provideCode string) ([]*biz.ModelProvider, error) {
	list, err := r.data.db.ModelProvider.Query().
		Where(
			modelprovider.ModelTypeEQ(modelType),
			modelprovider.ProviderCodeEQ(provideCode),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelProvider, len(list))
	for i, item := range list {
		result[i] = r.entityToBizModelProvider(item)
	}

	return result, nil
}

// entityToBizModelProvider 将ent实体转换为biz实体
func (r *modelProviderRepo) entityToBizModelProvider(entity *ent.ModelProvider) *biz.ModelProvider {
	return &biz.ModelProvider{
		ID:           entity.ID,
		ModelType:    entity.ModelType,
		ProviderCode: entity.ProviderCode,
		Name:         entity.Name,
		Fields:       entity.Fields,
		Sort:         entity.Sort,
		Creator:      entity.Creator,
		CreateDate:   entity.CreateDate,
		Updater:      entity.Updater,
		UpdateDate:   entity.UpdateDate,
	}
}

// GetListPage 分页查询模型供应器
func (r *modelProviderRepo) GetListPage(ctx context.Context, params *biz.ListModelProviderParams, page *kit.PageRequest) ([]*biz.ModelProvider, error) {
	query := r.buildModelProviderQuery(params, page)
	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelProvider, len(list))
	for i, item := range list {
		result[i] = r.entityToBizModelProvider(item)
	}

	return result, nil
}

// TotalModelProviders 获取模型供应器总数
func (r *modelProviderRepo) TotalModelProviders(ctx context.Context, params *biz.ListModelProviderParams) (int, error) {
	query := r.data.db.ModelProvider.Query()
	return r.applyModelProviderFilters(query, params).Count(ctx)
}

// buildModelProviderQuery 构建查询
func (r *modelProviderRepo) buildModelProviderQuery(params *biz.ListModelProviderParams, page *kit.PageRequest) *ent.ModelProviderQuery {
	query := r.data.db.ModelProvider.Query()
	query = r.applyModelProviderFilters(query, params)

	// 排序规则：按model_type和sort字段升序排序
	query = query.Order(ent.Asc(modelprovider.FieldModelType), ent.Asc(modelprovider.FieldSort))

	// 应用分页
	if page != nil {
		pageNo, _ := page.GetPageNo()
		pageSize := page.GetPageSize()
		if pageSize > 0 {
			query = query.Limit(pageSize)
			if pageNo > 0 {
				query = query.Offset((pageNo - 1) * pageSize)
			}
		}
	}

	return query
}

// applyModelProviderFilters 应用过滤条件
func (r *modelProviderRepo) applyModelProviderFilters(query *ent.ModelProviderQuery, params *biz.ListModelProviderParams) *ent.ModelProviderQuery {
	if params == nil {
		return query
	}

	// modelType精确匹配
	if params.ModelType != nil {
		query = query.Where(modelprovider.ModelTypeEQ(params.ModelType.GetValue()))
	}

	// name模糊查询name或provider_code字段（使用OR条件）
	if params.Name != nil {
		query = query.Where(
			modelprovider.Or(
				modelprovider.NameContains(params.Name.GetValue()),
				modelprovider.ProviderCodeContains(params.Name.GetValue()),
			),
		)
	}

	return query
}

// GetModelProviderByID 根据ID获取模型供应器
func (r *modelProviderRepo) GetModelProviderByID(ctx context.Context, id string) (*biz.ModelProvider, error) {
	provider, err := r.data.db.ModelProvider.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.entityToBizModelProvider(provider), nil
}

// CreateModelProvider 创建模型供应器
func (r *modelProviderRepo) CreateModelProvider(ctx context.Context, provider *biz.ModelProvider) (*biz.ModelProvider, error) {
	create := r.data.db.ModelProvider.Create().
		SetID(provider.ID).
		SetModelType(provider.ModelType).
		SetProviderCode(provider.ProviderCode).
		SetName(provider.Name).
		SetFields(provider.Fields).
		SetSort(provider.Sort).
		SetCreateDate(provider.CreateDate).
		SetUpdateDate(provider.UpdateDate)

	if provider.Creator > 0 {
		create.SetCreator(provider.Creator)
	}
	if provider.Updater > 0 {
		create.SetUpdater(provider.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.entityToBizModelProvider(entity), nil
}

// UpdateModelProvider 更新模型供应器
func (r *modelProviderRepo) UpdateModelProvider(ctx context.Context, provider *biz.ModelProvider) error {
	update := r.data.db.ModelProvider.UpdateOneID(provider.ID).
		SetModelType(provider.ModelType).
		SetProviderCode(provider.ProviderCode).
		SetName(provider.Name).
		SetFields(provider.Fields).
		SetSort(provider.Sort).
		SetUpdateDate(provider.UpdateDate)

	if provider.Updater > 0 {
		update.SetUpdater(provider.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteModelProvider 批量删除模型供应器
func (r *modelProviderRepo) DeleteModelProvider(ctx context.Context, ids []string) error {
	_, err := r.data.db.ModelProvider.Delete().
		Where(modelprovider.IDIn(ids...)).
		Exec(ctx)
	return err
}

// GetPluginList 获取Plugin类型的供应器列表
func (r *modelProviderRepo) GetPluginList(ctx context.Context) ([]*biz.ModelProvider, error) {
	list, err := r.data.db.ModelProvider.Query().
		Where(modelprovider.ModelTypeEQ("Plugin")).
		Order(ent.Asc(modelprovider.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.ModelProvider, len(list))
	for i, item := range list {
		result[i] = r.entityToBizModelProvider(item)
	}

	return result, nil
}

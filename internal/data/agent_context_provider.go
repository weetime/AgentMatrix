package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/agentcontextprovider"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type agentContextProviderRepo struct {
	data *Data
	log  *log.Helper
}

// NewAgentContextProviderRepo 初始化 AgentContextProvider Repo
func NewAgentContextProviderRepo(data *Data, logger log.Logger) biz.AgentContextProviderRepo {
	return &agentContextProviderRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/agent_context_provider")),
	}
}

// GetByAgentId 根据智能体ID获取上下文源配置
func (r *agentContextProviderRepo) GetByAgentId(ctx context.Context, agentId string) (*biz.AgentContextProvider, error) {
	entity, err := r.data.db.AgentContextProvider.Query().
		Where(agentcontextprovider.AgentIDEQ(agentId)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.entityToBiz(entity)
}

// SaveOrUpdateByAgentId 保存或更新上下文源配置
func (r *agentContextProviderRepo) SaveOrUpdateByAgentId(ctx context.Context, entity *biz.AgentContextProvider) error {
	// 检查是否存在
	existing, err := r.data.db.AgentContextProvider.Query().
		Where(agentcontextprovider.AgentIDEQ(entity.AgentID)).
		Only(ctx)

	// 序列化 context_providers 为 JSON
	contextProvidersJSON := ""
	if entity.ContextProviders != nil && len(entity.ContextProviders) > 0 {
		jsonBytes, err := json.Marshal(entity.ContextProviders)
		if err != nil {
			return err
		}
		contextProvidersJSON = string(jsonBytes)
	}

	if err != nil {
		if ent.IsNotFound(err) {
			// 不存在，创建新记录
			create := r.data.db.AgentContextProvider.Create().
				SetID(uuid.New().String()[:32]).
				SetAgentID(entity.AgentID).
				SetCreatedAt(time.Now()).
				SetUpdatedAt(time.Now())

			if contextProvidersJSON != "" {
				create = create.SetContextProviders(contextProvidersJSON)
			}
			if entity.Creator != 0 {
				create = create.SetCreator(entity.Creator)
			} else if entity.Updater != 0 {
				create = create.SetCreator(entity.Updater)
			}
			if entity.Updater != 0 {
				create = create.SetUpdater(entity.Updater)
			}

			_, err = create.Save(ctx)
			return err
		}
		return err
	}

	// 存在，更新记录
	update := r.data.db.AgentContextProvider.UpdateOneID(existing.ID).
		SetUpdatedAt(time.Now())

	if contextProvidersJSON != "" {
		update = update.SetContextProviders(contextProvidersJSON)
	} else {
		update = update.SetContextProviders("")
	}
	if entity.Updater != 0 {
		update = update.SetUpdater(entity.Updater)
	}

	_, err = update.Save(ctx)
	return err
}

// DeleteByAgentId 根据智能体ID删除上下文源配置
func (r *agentContextProviderRepo) DeleteByAgentId(ctx context.Context, agentId string) error {
	_, err := r.data.db.AgentContextProvider.Delete().
		Where(agentcontextprovider.AgentIDEQ(agentId)).
		Exec(ctx)
	return err
}

// entityToBiz 将 Ent 实体转换为业务实体
func (r *agentContextProviderRepo) entityToBiz(entity *ent.AgentContextProvider) (*biz.AgentContextProvider, error) {
	result := &biz.AgentContextProvider{
		ID:        entity.ID,
		AgentID:   entity.AgentID,
		Creator:   entity.Creator,
		CreatedAt: entity.CreatedAt,
		Updater:   entity.Updater,
		UpdatedAt: entity.UpdatedAt,
	}

	// 反序列化 context_providers JSON
	if entity.ContextProviders != "" {
		var providers []*biz.ContextProviderDTO
		if err := json.Unmarshal([]byte(entity.ContextProviders), &providers); err != nil {
			r.log.Warnf("Failed to unmarshal context_providers JSON: %v", err)
			result.ContextProviders = []*biz.ContextProviderDTO{}
		} else {
			result.ContextProviders = providers
		}
	} else {
		result.ContextProviders = []*biz.ContextProviderDTO{}
	}

	return result, nil
}

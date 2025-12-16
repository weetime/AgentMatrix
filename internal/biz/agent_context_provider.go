package biz

import (
	"context"
	"time"

	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// ContextProviderDTO 上下文源配置DTO
type ContextProviderDTO struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// AgentContextProvider 智能体上下文源配置业务实体
type AgentContextProvider struct {
	ID               string
	AgentID          string
	ContextProviders []*ContextProviderDTO
	Creator          int64
	CreatedAt        time.Time
	Updater          int64
	UpdatedAt        time.Time
}

// AgentContextProviderRepo 智能体上下文源配置数据访问接口
type AgentContextProviderRepo interface {
	GetByAgentId(ctx context.Context, agentId string) (*AgentContextProvider, error)
	SaveOrUpdateByAgentId(ctx context.Context, entity *AgentContextProvider) error
	DeleteByAgentId(ctx context.Context, agentId string) error
}

// AgentContextProviderUsecase 智能体上下文源配置业务逻辑
type AgentContextProviderUsecase struct {
	repo        AgentContextProviderRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewAgentContextProviderUsecase 创建智能体上下文源配置用例
func NewAgentContextProviderUsecase(
	repo AgentContextProviderRepo,
	logger log.Logger,
) *AgentContextProviderUsecase {
	return &AgentContextProviderUsecase{
		repo:        repo,
		handleError: cerrors.NewHandleError(logger),
		log:         log.NewHelper(log.With(logger, "module", "agent-matrix-service/biz/agent_context_provider")),
	}
}

// GetByAgentId 根据智能体ID获取上下文源配置
func (uc *AgentContextProviderUsecase) GetByAgentId(ctx context.Context, agentId string) (*AgentContextProvider, error) {
	return uc.repo.GetByAgentId(ctx, agentId)
}

// SaveOrUpdateByAgentId 保存或更新上下文源配置
func (uc *AgentContextProviderUsecase) SaveOrUpdateByAgentId(ctx context.Context, agentId string, providers []*ContextProviderDTO, userId int64) error {
	entity := &AgentContextProvider{
		AgentID:          agentId,
		ContextProviders: providers,
		Updater:          userId,
		UpdatedAt:        time.Now(),
	}
	return uc.repo.SaveOrUpdateByAgentId(ctx, entity)
}

// DeleteByAgentId 根据智能体ID删除上下文源配置
func (uc *AgentContextProviderUsecase) DeleteByAgentId(ctx context.Context, agentId string) error {
	return uc.repo.DeleteByAgentId(ctx, agentId)
}

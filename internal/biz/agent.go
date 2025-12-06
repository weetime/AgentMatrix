package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Agent 智能体实体
type Agent struct {
	ID              string
	UserID          int64
	AgentCode       string
	AgentName       string
	ASRModelID      string
	VADModelID      string
	LLMModelID      string
	VLLMModelID     string
	TTSModelID      string
	TTSVoiceID      string
	MemModelID      string
	IntentModelID   string
	ChatHistoryConf int8
	SystemPrompt    string
	SummaryMemory   string
	LangCode        string
	Language        string
	Sort            int8
	Creator         int64
	CreatedAt       time.Time
	Updater         int64
	UpdatedAt       time.Time
}

// AgentDTO 智能体列表项
type AgentDTO struct {
	ID              string
	AgentName       string
	TTSModelName    string
	TTSVoiceName    string
	LLMModelName    string
	VLLMModelName   string
	MemModelID      string
	SystemPrompt    string
	SummaryMemory   string
	LastConnectedAt string
	DeviceCount     int32
}

// AgentPluginMapping 插件映射
type AgentPluginMapping struct {
	ID           int64
	AgentID      string
	PluginID     string
	ParamInfo    string
	ProviderCode string
}

// AgentTemplate 智能体模板
type AgentTemplate struct {
	ID              string
	AgentCode       string
	AgentName       string
	ASRModelID      string
	VADModelID      string
	LLMModelID      string
	VLLMModelID     string
	TTSModelID      string
	TTSVoiceID      string
	MemModelID      string
	IntentModelID   string
	ChatHistoryConf int8
	SystemPrompt    string
	SummaryMemory   string
	LangCode        string
	Language        string
	Sort            int8
}

// AgentTemplateVO 智能体模板VO（包含模型名称）
type AgentTemplateVO struct {
	ID              string
	AgentCode       string
	AgentName       string
	ASRModelID      string
	VADModelID      string
	LLMModelID      string
	VLLMModelID     string
	TTSModelID      string
	TTSVoiceID      string
	MemModelID      string
	IntentModelID   string
	ChatHistoryConf int8
	SystemPrompt    string
	SummaryMemory   string
	LangCode        string
	Language        string
	Sort            int8
	TTSModelName    string
	LLMModelName    string
}

// ListAgentTemplateParams 模板查询参数
type ListAgentTemplateParams struct {
	AgentName *string // 模板名称，模糊查询
}

// AgentChatSession 智能体会话
type AgentChatSession struct {
	SessionID string
	CreatedAt time.Time
	ChatCount int32
}

// AgentChatHistory 智能体聊天记录
type AgentChatHistory struct {
	ID         int64
	MacAddress *string
	AgentID    *string
	SessionID  *string
	ChatType   int8
	Content    *string
	AudioID    *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentChatHistoryUserVO 用户聊天记录VO
type AgentChatHistoryUserVO struct {
	Content string
	AudioID string
}

// AgentRepo 智能体数据访问接口
type AgentRepo interface {
	ListUserAgents(ctx context.Context, userId int64) ([]*AgentDTO, error)
	ListAllAgents(ctx context.Context, page *kit.PageRequest) ([]*Agent, int, error)
	GetAgentByID(ctx context.Context, id string) (*Agent, []*AgentPluginMapping, error)
	CreateAgent(ctx context.Context, agent *Agent) (*Agent, error)
	UpdateAgent(ctx context.Context, agent *Agent) error
	DeleteAgent(ctx context.Context, id string) error
	GetAgentTemplateList(ctx context.Context) ([]*AgentTemplate, error)
	GetAgentTemplatePage(ctx context.Context, params *ListAgentTemplateParams, page *kit.PageRequest) ([]*AgentTemplate, int, error)
	GetAgentTemplateByID(ctx context.Context, id string) (*AgentTemplate, error)
	CreateAgentTemplate(ctx context.Context, template *AgentTemplate) (*AgentTemplate, error)
	UpdateAgentTemplate(ctx context.Context, template *AgentTemplate) error
	DeleteAgentTemplate(ctx context.Context, id string) error
	BatchDeleteAgentTemplates(ctx context.Context, ids []string) error
	GetNextAvailableSort(ctx context.Context) (int8, error)
	ReorderTemplatesAfterDelete(ctx context.Context, deletedSort int8) error
	GetSessionsByAgentID(ctx context.Context, agentId string, page *kit.PageRequest) ([]*AgentChatSession, int, error)
	GetChatHistoryBySessionID(ctx context.Context, agentId, sessionId string) ([]*AgentChatHistory, error)
	GetRecentFiftyUserChats(ctx context.Context, agentId string) ([]*AgentChatHistoryUserVO, error)
	GetContentByAudioID(ctx context.Context, audioId string) (string, error)
	GetAudioByID(ctx context.Context, audioId string) ([]byte, error)
	SaveAudio(ctx context.Context, audioId string, audioData []byte) error
	DeleteChatHistoryByAgentID(ctx context.Context, agentId string) error
	DeleteAudioByAgentID(ctx context.Context, agentId string) error
	GetDeviceCountByAgentID(ctx context.Context, agentId string) (int, error)
	GetDefaultAgentByMacAddress(ctx context.Context, macAddress string) (*Agent, error)
}

// AgentUsecase 智能体业务逻辑
type AgentUsecase struct {
	repo        AgentRepo
	redisClient *kit.RedisClient
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewAgentUsecase 创建智能体用例
func NewAgentUsecase(
	repo AgentRepo,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *AgentUsecase {
	return &AgentUsecase{
		repo:        repo,
		redisClient: redisClient,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// GenerateAgentID 生成智能体ID（UUID去掉横线）
func (uc *AgentUsecase) GenerateAgentID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenerateAgentCode 生成智能体编码
func (uc *AgentUsecase) GenerateAgentCode() string {
	return fmt.Sprintf("AGT_%d", time.Now().UnixMilli())
}

// ListUserAgents 获取用户智能体列表
func (uc *AgentUsecase) ListUserAgents(ctx context.Context, userId int64) ([]*AgentDTO, error) {
	return uc.repo.ListUserAgents(ctx, userId)
}

// ListAllAgents 管理员列表（分页）
func (uc *AgentUsecase) ListAllAgents(ctx context.Context, page *kit.PageRequest) ([]*Agent, int, error) {
	return uc.repo.ListAllAgents(ctx, page)
}

// GetAgentByID 获取详情
func (uc *AgentUsecase) GetAgentByID(ctx context.Context, id string) (*Agent, []*AgentPluginMapping, error) {
	return uc.repo.GetAgentByID(ctx, id)
}

// CreateAgent 创建智能体
func (uc *AgentUsecase) CreateAgent(ctx context.Context, agentName string, userId int64) (*Agent, error) {
	agent := &Agent{
		ID:              uc.GenerateAgentID(),
		UserID:          userId,
		AgentCode:       uc.GenerateAgentCode(),
		AgentName:       agentName,
		ChatHistoryConf: 0,
		Sort:            0,
		Creator:         userId,
	}

	return uc.repo.CreateAgent(ctx, agent)
}

// UpdateAgent 更新智能体
func (uc *AgentUsecase) UpdateAgent(ctx context.Context, agent *Agent) error {
	// 检查智能体是否存在
	existing, _, err := uc.repo.GetAgentByID(ctx, agent.ID)
	if err != nil {
		return fmt.Errorf("智能体不存在: %w", err)
	}

	// 只更新提供的字段
	if agent.AgentCode != "" {
		existing.AgentCode = agent.AgentCode
	}
	if agent.AgentName != "" {
		existing.AgentName = agent.AgentName
	}
	if agent.ASRModelID != "" {
		existing.ASRModelID = agent.ASRModelID
	}
	if agent.VADModelID != "" {
		existing.VADModelID = agent.VADModelID
	}
	if agent.LLMModelID != "" {
		existing.LLMModelID = agent.LLMModelID
	}
	if agent.VLLMModelID != "" {
		existing.VLLMModelID = agent.VLLMModelID
	}
	if agent.TTSModelID != "" {
		existing.TTSModelID = agent.TTSModelID
	}
	if agent.TTSVoiceID != "" {
		existing.TTSVoiceID = agent.TTSVoiceID
	}
	if agent.MemModelID != "" {
		existing.MemModelID = agent.MemModelID
	}
	if agent.IntentModelID != "" {
		existing.IntentModelID = agent.IntentModelID
	}
	if agent.SystemPrompt != "" {
		existing.SystemPrompt = agent.SystemPrompt
	}
	if agent.SummaryMemory != "" {
		existing.SummaryMemory = agent.SummaryMemory
	}
	if agent.LangCode != "" {
		existing.LangCode = agent.LangCode
	}
	if agent.Language != "" {
		existing.Language = agent.Language
	}
	if agent.Updater > 0 {
		existing.Updater = agent.Updater
	}

	return uc.repo.UpdateAgent(ctx, existing)
}

// UpdateAgentMemoryByMacAddress 根据设备更新智能体记忆
func (uc *AgentUsecase) UpdateAgentMemoryByMacAddress(ctx context.Context, macAddress, summaryMemory string) error {
	// 根据 MAC 地址获取智能体
	agent, err := uc.repo.GetDefaultAgentByMacAddress(ctx, macAddress)
	if err != nil {
		return fmt.Errorf("设备不存在或未关联智能体: %w", err)
	}

	agent.SummaryMemory = summaryMemory
	return uc.repo.UpdateAgent(ctx, agent)
}

// DeleteAgent 删除智能体（级联删除）
func (uc *AgentUsecase) DeleteAgent(ctx context.Context, id string) error {
	// 删除聊天记录（文本+音频）
	if err := uc.repo.DeleteChatHistoryByAgentID(ctx, id); err != nil {
		uc.log.Warn("Failed to delete chat history", "agentId", id, "error", err)
	}

	// 删除音频数据
	if err := uc.repo.DeleteAudioByAgentID(ctx, id); err != nil {
		uc.log.Warn("Failed to delete audio", "agentId", id, "error", err)
	}

	// TODO: 删除关联的设备（需要 DeviceService）
	// TODO: 删除关联的插件映射（需要 AgentPluginMappingService）

	// 删除智能体
	return uc.repo.DeleteAgent(ctx, id)
}

// GetAgentTemplateList 获取模板列表
func (uc *AgentUsecase) GetAgentTemplateList(ctx context.Context) ([]*AgentTemplate, error) {
	return uc.repo.GetAgentTemplateList(ctx)
}

// GetAgentTemplatePage 分页查询模板
func (uc *AgentUsecase) GetAgentTemplatePage(ctx context.Context, params *ListAgentTemplateParams, page *kit.PageRequest) ([]*AgentTemplate, int, error) {
	return uc.repo.GetAgentTemplatePage(ctx, params, page)
}

// GetAgentTemplateByID 获取模板详情
func (uc *AgentUsecase) GetAgentTemplateByID(ctx context.Context, id string) (*AgentTemplate, error) {
	return uc.repo.GetAgentTemplateByID(ctx, id)
}

// CreateAgentTemplate 创建模板
func (uc *AgentUsecase) CreateAgentTemplate(ctx context.Context, template *AgentTemplate) (*AgentTemplate, error) {
	// 获取下一个可用的排序值
	sort, err := uc.repo.GetNextAvailableSort(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取排序值失败: %w", err)
	}
	template.Sort = sort

	// 生成ID
	if template.ID == "" {
		template.ID = uc.GenerateAgentID()
	}

	return uc.repo.CreateAgentTemplate(ctx, template)
}

// UpdateAgentTemplate 更新模板
func (uc *AgentUsecase) UpdateAgentTemplate(ctx context.Context, template *AgentTemplate) error {
	// 检查模板是否存在
	existing, err := uc.repo.GetAgentTemplateByID(ctx, template.ID)
	if err != nil {
		return fmt.Errorf("模板不存在: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("模板不存在")
	}

	// 只更新提供的字段
	if template.AgentCode != "" {
		existing.AgentCode = template.AgentCode
	}
	if template.AgentName != "" {
		existing.AgentName = template.AgentName
	}
	if template.ASRModelID != "" {
		existing.ASRModelID = template.ASRModelID
	}
	if template.VADModelID != "" {
		existing.VADModelID = template.VADModelID
	}
	if template.LLMModelID != "" {
		existing.LLMModelID = template.LLMModelID
	}
	if template.VLLMModelID != "" {
		existing.VLLMModelID = template.VLLMModelID
	}
	if template.TTSModelID != "" {
		existing.TTSModelID = template.TTSModelID
	}
	if template.TTSVoiceID != "" {
		existing.TTSVoiceID = template.TTSVoiceID
	}
	if template.MemModelID != "" {
		existing.MemModelID = template.MemModelID
	}
	if template.IntentModelID != "" {
		existing.IntentModelID = template.IntentModelID
	}
	if template.SystemPrompt != "" {
		existing.SystemPrompt = template.SystemPrompt
	}
	if template.SummaryMemory != "" {
		existing.SummaryMemory = template.SummaryMemory
	}
	if template.LangCode != "" {
		existing.LangCode = template.LangCode
	}
	if template.Language != "" {
		existing.Language = template.Language
	}

	return uc.repo.UpdateAgentTemplate(ctx, existing)
}

// DeleteAgentTemplate 删除模板
func (uc *AgentUsecase) DeleteAgentTemplate(ctx context.Context, id string) error {
	// 先查询要删除的模板信息，获取其排序值
	template, err := uc.repo.GetAgentTemplateByID(ctx, id)
	if err != nil {
		return fmt.Errorf("模板不存在: %w", err)
	}
	if template == nil {
		return fmt.Errorf("模板不存在")
	}

	deletedSort := template.Sort

	// 执行删除操作
	err = uc.repo.DeleteAgentTemplate(ctx, id)
	if err != nil {
		return err
	}

	// 删除成功后，重新排序剩余模板
	return uc.repo.ReorderTemplatesAfterDelete(ctx, deletedSort)
}

// BatchDeleteAgentTemplates 批量删除模板
func (uc *AgentUsecase) BatchDeleteAgentTemplates(ctx context.Context, ids []string) error {
	return uc.repo.BatchDeleteAgentTemplates(ctx, ids)
}

// GetAgentSessions 获取智能体会话列表
func (uc *AgentUsecase) GetAgentSessions(ctx context.Context, agentId string, page *kit.PageRequest) ([]*AgentChatSession, int, error) {
	return uc.repo.GetSessionsByAgentID(ctx, agentId, page)
}

// GetAgentChatHistory 获取智能体聊天记录
func (uc *AgentUsecase) GetAgentChatHistory(ctx context.Context, agentId, sessionId string) ([]*AgentChatHistory, error) {
	return uc.repo.GetChatHistoryBySessionID(ctx, agentId, sessionId)
}

// GetRecentFiftyUserChats 获取最近50条用户聊天
func (uc *AgentUsecase) GetRecentFiftyUserChats(ctx context.Context, agentId string) ([]*AgentChatHistoryUserVO, error) {
	return uc.repo.GetRecentFiftyUserChats(ctx, agentId)
}

// GetContentByAudioID 根据音频ID获取内容
func (uc *AgentUsecase) GetContentByAudioID(ctx context.Context, audioId string) (string, error) {
	return uc.repo.GetContentByAudioID(ctx, audioId)
}

// GetAudioDownloadID 获取音频下载ID（生成UUID临时链接）
func (uc *AgentUsecase) GetAudioDownloadID(ctx context.Context, audioId string) (string, error) {
	// 检查音频是否存在
	_, err := uc.repo.GetAudioByID(ctx, audioId)
	if err != nil {
		return "", fmt.Errorf("音频不存在: %w", err)
	}

	// 生成UUID
	uuidStr := uuid.New().String()

	// 存储到Redis（24小时过期）
	key := fmt.Sprintf("agent:audio:%s", uuidStr)
	if uc.redisClient != nil {
		if err := uc.redisClient.Set(ctx, key, audioId, 24*3600); err != nil {
			uc.log.Warn("Failed to save audio download ID to redis", "error", err)
		}
	}

	return uuidStr, nil
}

// PlayAudio 播放音频（通过UUID获取音频数据）
func (uc *AgentUsecase) PlayAudio(ctx context.Context, uuidStr string) ([]byte, error) {
	// 从Redis获取audioId
	key := fmt.Sprintf("agent:audio:%s", uuidStr)
	var audioId string
	if uc.redisClient != nil {
		val, err := uc.redisClient.Get(ctx, key)
		if err != nil || val == "" {
			return nil, fmt.Errorf("下载链接已过期或不存在")
		}
		audioId = val
		// 使用后删除
		uc.redisClient.Delete(ctx, key)
	} else {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	// 获取音频数据
	return uc.repo.GetAudioByID(ctx, audioId)
}

// CheckAgentPermission 检查用户是否有权限访问智能体
func (uc *AgentUsecase) CheckAgentPermission(ctx context.Context, agentId string, userId int64, isSuperAdmin bool) (bool, error) {
	if isSuperAdmin {
		return true, nil
	}

	agent, _, err := uc.repo.GetAgentByID(ctx, agentId)
	if err != nil {
		return false, err
	}

	return agent.UserID == userId, nil
}

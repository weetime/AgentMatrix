package biz

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/weetime/agent-matrix/internal/constant"
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
	TTSModelID      string // 内部使用，用于查询模型名称
	TTSModelName    string
	TTSVoiceID      string // 内部使用，用于查询音色名称
	TTSVoiceName    string
	LLMModelID      string // 内部使用，用于查询模型名称
	LLMModelName    string
	VLLMModelID     string // 内部使用，用于查询模型名称
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
	SaveChatHistory(ctx context.Context, history *AgentChatHistory) error
	GetRecentFiftyUserChats(ctx context.Context, agentId string) ([]*AgentChatHistoryUserVO, error)
	GetContentByAudioID(ctx context.Context, audioId string) (string, error)
	GetAudioByID(ctx context.Context, audioId string) ([]byte, error)
	SaveAudio(ctx context.Context, audioId string, audioData []byte) error
	DeleteChatHistoryByAgentID(ctx context.Context, agentId string) error
	DeleteAudioByAgentID(ctx context.Context, agentId string) error
	GetDeviceCountByAgentID(ctx context.Context, agentId string) (int, error)
	GetLatestLastConnectionTimeByAgentID(ctx context.Context, agentId string) (*time.Time, error)
	GetDefaultAgentByMacAddress(ctx context.Context, macAddress string) (*Agent, error)
	IsAudioOwnedByAgent(ctx context.Context, audioId, agentId string) (bool, error)
	DeleteAgentsByUserId(ctx context.Context, userId int64) error
}

// AgentUsecase 智能体业务逻辑
type AgentUsecase struct {
	repo            AgentRepo
	configUsecase   *ConfigUsecase
	modelUsecase    *ModelUsecase
	ttsVoiceUsecase *TtsVoiceUsecase
	redisClient     *kit.RedisClient
	handleError     *cerrors.HandleError
	log             *log.Helper
}

// NewAgentUsecase 创建智能体用例
func NewAgentUsecase(
	repo AgentRepo,
	configUsecase *ConfigUsecase,
	modelUsecase *ModelUsecase,
	ttsVoiceUsecase *TtsVoiceUsecase,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *AgentUsecase {
	return &AgentUsecase{
		repo:            repo,
		configUsecase:   configUsecase,
		modelUsecase:    modelUsecase,
		ttsVoiceUsecase: ttsVoiceUsecase,
		redisClient:     redisClient,
		handleError:     cerrors.NewHandleError(logger),
		log:             kit.LogHelper(logger),
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
	agents, err := uc.repo.ListUserAgents(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 关联查询模型名称、音色名称和最后连接时间
	for _, agent := range agents {
		// 获取 TTS 模型名称
		if agent.TTSModelID != "" {
			if config, err := uc.modelUsecase.GetModelConfigByID(ctx, agent.TTSModelID); err == nil && config != nil {
				agent.TTSModelName = config.ModelName
			}
		}

		// 获取 LLM 模型名称
		if agent.LLMModelID != "" {
			if config, err := uc.modelUsecase.GetModelConfigByID(ctx, agent.LLMModelID); err == nil && config != nil {
				agent.LLMModelName = config.ModelName
			}
		}

		// 获取 VLLM 模型名称
		if agent.VLLMModelID != "" {
			if config, err := uc.modelUsecase.GetModelConfigByID(ctx, agent.VLLMModelID); err == nil && config != nil {
				agent.VLLMModelName = config.ModelName
			}
		}

		// 获取 TTS 音色名称
		if agent.TTSVoiceID != "" {
			if voice, err := uc.ttsVoiceUsecase.GetTtsVoiceByID(ctx, agent.TTSVoiceID); err == nil && voice != nil {
				agent.TTSVoiceName = voice.Name
			}
		}

		// 获取最后连接时间
		if lastTime, err := uc.repo.GetLatestLastConnectionTimeByAgentID(ctx, agent.ID); err == nil && lastTime != nil {
			agent.LastConnectedAt = lastTime.Format("2006-01-02 15:04:05")
		}
	}

	return agents, nil
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

// ReportChatHistoryRequest 聊天上报请求
type ReportChatHistoryRequest struct {
	MacAddress  string
	SessionID   string
	ChatType    int8
	Content     string
	AudioBase64 *string
	ReportTime  *int64 // 十位时间戳，nil时使用当前时间
}

// ReportChatHistory 处理聊天记录上报
func (uc *AgentUsecase) ReportChatHistory(ctx context.Context, req *ReportChatHistoryRequest) (bool, error) {
	// 根据 MAC 地址获取默认智能体
	agent, err := uc.repo.GetDefaultAgentByMacAddress(ctx, req.MacAddress)
	if err != nil {
		return false, fmt.Errorf("获取智能体失败: %w", err)
	}
	if agent == nil {
		uc.log.Warnf("MAC地址 %s 未找到对应的智能体", req.MacAddress)
		return false, nil
	}

	// 确定保存策略
	chatHistoryConf := agent.ChatHistoryConf
	var audioID *string

	// 如果需要保存音频
	if chatHistoryConf == constant.ChatHistoryConfRecordTextAudio {
		if req.AudioBase64 != nil && *req.AudioBase64 != "" {
			// Base64 解码音频数据
			audioData, err := decodeBase64(*req.AudioBase64)
			if err != nil {
				uc.log.Errorf("音频数据解码失败: %v", err)
				return false, fmt.Errorf("音频数据解码失败: %w", err)
			}

			// 保存音频并获取 audioID
			audioIDStr := uuid.New().String()
			if err := uc.repo.SaveAudio(ctx, audioIDStr, audioData); err != nil {
				uc.log.Errorf("音频数据保存失败: %v", err)
				return false, fmt.Errorf("音频数据保存失败: %w", err)
			}
			audioID = &audioIDStr
			uc.log.Infof("音频数据保存成功，audioId=%s", audioIDStr)
		}
	}

	// 如果需要保存文本
	if chatHistoryConf == constant.ChatHistoryConfRecordText || chatHistoryConf == constant.ChatHistoryConfRecordTextAudio {
		// 确定创建时间
		var createdAt time.Time
		if req.ReportTime != nil {
			createdAt = time.Unix(*req.ReportTime, 0)
		} else {
			createdAt = time.Now()
		}

		// 构建聊天记录实体
		macAddress := req.MacAddress
		agentID := agent.ID
		sessionID := req.SessionID
		content := req.Content

		history := &AgentChatHistory{
			MacAddress: &macAddress,
			AgentID:    &agentID,
			SessionID:  &sessionID,
			ChatType:   req.ChatType,
			Content:    &content,
			AudioID:    audioID,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}

		// 保存聊天记录
		if err := uc.repo.SaveChatHistory(ctx, history); err != nil {
			return false, fmt.Errorf("保存聊天记录失败: %w", err)
		}

		uc.log.Infof("设备 %s 对应智能体 %s 上报成功", req.MacAddress, agent.ID)
	}

	// 更新设备最后连接时间到 Redis
	if uc.redisClient != nil {
		key := fmt.Sprintf("agent:device:lastConnectedAt:%s", agent.ID)
		now := time.Now()
		if err := uc.redisClient.Set(ctx, key, now, 0); err != nil {
			uc.log.Warnf("更新设备最后连接时间到Redis失败: %v", err)
		}
	}

	// 更新设备最后连接时间（通过 DeviceUsecase）
	// 注意：这里需要 DeviceUsecase，但当前 AgentUsecase 没有这个依赖
	// 可以考虑通过 repo 直接更新，或者添加 DeviceUsecase 依赖
	// 暂时先跳过设备更新，因为需要 DeviceUsecase

	return true, nil
}

// GetChatHistoryDownloadUrl 获取聊天记录下载链接
func (uc *AgentUsecase) GetChatHistoryDownloadUrl(ctx context.Context, agentId, sessionId string) (string, error) {
	// 生成 UUID
	uuidStr := uuid.New().String()

	// 存储到 Redis（key: agent:chat:history:{uuid}, value: agentId:sessionId）
	key := fmt.Sprintf("agent:chat:history:%s", uuidStr)
	value := fmt.Sprintf("%s:%s", agentId, sessionId)
	if uc.redisClient != nil {
		// 设置24小时过期
		if err := uc.redisClient.Set(ctx, key, value, 24*time.Hour); err != nil {
			return "", fmt.Errorf("保存下载链接到Redis失败: %w", err)
		}
	} else {
		return "", fmt.Errorf("Redis客户端未初始化")
	}

	return uuidStr, nil
}

// DownloadChatHistoryRequest 下载聊天记录请求
type DownloadChatHistoryRequest struct {
	UUID            string
	IncludePrevious bool // 是否包含前20条会话
}

// DownloadChatHistoryResponse 下载聊天记录响应
type DownloadChatHistoryResponse struct {
	Content []byte
}

// DownloadChatHistory 下载聊天记录
func (uc *AgentUsecase) DownloadChatHistory(ctx context.Context, uuid string, includePrevious bool) ([]byte, error) {
	// 从 Redis 获取 agentId 和 sessionId
	key := fmt.Sprintf("agent:chat:history:%s", uuid)
	var agentSessionInfo string
	if uc.redisClient != nil {
		val, err := uc.redisClient.Get(ctx, key)
		if err != nil || val == "" {
			return nil, fmt.Errorf("下载链接已过期或无效")
		}
		agentSessionInfo = val
	} else {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	// 解析 agentId 和 sessionId
	parts := strings.Split(agentSessionInfo, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("下载链接无效")
	}
	agentId := parts[0]
	sessionId := parts[1]

	var sessionIds []string
	if includePrevious {
		// 获取所有会话列表
		page := &kit.PageRequest{}
		page.SetPageNo(1)
		page.SetPageSize(1000)
		sessions, _, err := uc.repo.GetSessionsByAgentID(ctx, agentId, page)
		if err != nil {
			return nil, fmt.Errorf("获取会话列表失败: %w", err)
		}

		// 找到当前会话在列表中的位置
		currentIndex := -1
		for i, session := range sessions {
			if session.SessionID == sessionId {
				currentIndex = i
				break
			}
		}

		// 从当前会话开始，向后取最多20条会话（包括当前会话）
		if currentIndex != -1 {
			endIndex := currentIndex + 20
			if endIndex > len(sessions) {
				endIndex = len(sessions)
			}
			for i := currentIndex; i < endIndex; i++ {
				sessionIds = append(sessionIds, sessions[i].SessionID)
			}
		} else {
			// 如果没找到，至少下载当前会话
			sessionIds = []string{sessionId}
		}
	} else {
		// 只下载当前会话
		sessionIds = []string{sessionId}
	}

	// 生成文本内容
	var content strings.Builder
	for i, sid := range sessionIds {
		// 获取该会话的所有聊天记录
		histories, err := uc.repo.GetChatHistoryBySessionID(ctx, agentId, sid)
		if err != nil {
			return nil, fmt.Errorf("获取聊天记录失败: %w", err)
		}

		// 从聊天记录中获取第一条消息的创建时间作为会话时间
		if len(histories) > 0 {
			firstMessageTime := histories[0].CreatedAt
			sessionTimeStr := formatDateTime(firstMessageTime)
			content.WriteString(sessionTimeStr)
			content.WriteString("\n")
		}

		// 写入每条消息
		for _, msg := range histories {
			role := "用户"
			direction := ">>"
			if msg.ChatType == 2 {
				role = "智能体"
				direction = "<<"
			}
			messageTimeStr := formatDateTime(msg.CreatedAt)
			msgContent := ""
			if msg.Content != nil {
				msgContent = *msg.Content
			}
			line := fmt.Sprintf("[%s]-[%s]%s:%s\n", role, messageTimeStr, direction, msgContent)
			content.WriteString(line)
		}

		// 会话之间添加空行分隔
		if i < len(sessionIds)-1 {
			content.WriteString("\n")
		}
	}

	// 下载完成后删除 Redis key
	if uc.redisClient != nil {
		uc.redisClient.Delete(ctx, key)
	}

	return []byte(content.String()), nil
}

// decodeBase64 解码 Base64 字符串
func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// formatDateTime 格式化日期时间为 yyyy-MM-dd HH:mm:ss
func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetAgentMcpAccessAddress 获取智能体的MCP接入点地址
func (uc *AgentUsecase) GetAgentMcpAccessAddress(ctx context.Context, agentId string) (string, error) {
	// 获取MCP地址配置
	mcpUrl, err := uc.configUsecase.GetValue(ctx, "server.mcp_endpoint", true)
	if err != nil {
		return "", fmt.Errorf("获取MCP配置失败: %w", err)
	}

	if mcpUrl == "" || mcpUrl == "null" {
		return "", nil
	}

	// 解析URI
	parsedURL, err := parseURI(mcpUrl)
	if err != nil {
		return "", fmt.Errorf("mcp的地址存在错误，请进入参数管理修改mcp接入点地址: %w", err)
	}

	// 获取智能体mcp的url前缀
	agentMcpUrl := uc.getAgentMcpUrl(parsedURL)

	// 获取密钥
	key := uc.getSecretKey(parsedURL)

	// 获取加密的token
	encryptToken, err := uc.encryptToken(agentId, key)
	if err != nil {
		return "", fmt.Errorf("加密token失败: %w", err)
	}

	// 对token进行URL编码
	encodedToken := url.QueryEscape(encryptToken)

	// 返回智能体Mcp路径的格式
	agentMcpUrl = fmt.Sprintf("%s/mcp/?token=%s", agentMcpUrl, encodedToken)
	return agentMcpUrl, nil
}

// GetAgentMcpToolsList 获取智能体的MCP工具列表
func (uc *AgentUsecase) GetAgentMcpToolsList(ctx context.Context, agentId string) ([]string, error) {
	// 获取MCP地址
	wsUrl, err := uc.GetAgentMcpAccessAddress(ctx, agentId)
	if err != nil {
		return nil, err
	}

	if wsUrl == "" {
		return []string{}, nil
	}

	// 将 /mcp 替换为 /call
	wsUrl = strings.Replace(wsUrl, "/mcp/", "/call/", 1)

	// 创建WebSocket客户端
	client := kit.NewWebSocketClient(&kit.WebSocketClientConfig{
		URL:            wsUrl,
		ConnectTimeout: 8 * time.Second,
		MaxDuration:    10 * time.Second,
		BufferSize:     1024 * 1024,
		Logger:         uc.log,
	})

	// 确保连接关闭
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		uc.log.Warnf("WebSocket连接失败，智能体ID: %s, 错误: %v", agentId, err)
		return []string{}, nil
	}

	// 步骤1: 发送初始化消息并等待响应
	uc.log.Infof("发送MCP初始化消息，智能体ID: %s", agentId)
	if err := client.SendText(kit.GetInitializeJson()); err != nil {
		uc.log.Warnf("发送初始化消息失败: %v", err)
		return []string{}, nil
	}

	// 等待初始化响应 (id=1)
	initResponses := client.ListenForResponseWithoutClose(func(response string) bool {
		jsonMap, err := kit.ParseJsonResponse(response)
		if err != nil {
			return false
		}
		if id, ok := jsonMap["id"].(float64); ok && int(id) == 1 {
			// 检查是否有result字段，表示初始化成功
			_, hasResult := jsonMap["result"]
			_, hasError := jsonMap["error"]
			return hasResult && !hasError
		}
		return false
	})

	// 验证初始化响应
	initSucceeded := false
	for _, response := range initResponses {
		jsonMap, err := kit.ParseJsonResponse(response)
		if err != nil {
			continue
		}
		if id, ok := jsonMap["id"].(float64); ok && int(id) == 1 {
			if _, hasResult := jsonMap["result"]; hasResult {
				uc.log.Infof("MCP初始化成功，智能体ID: %s", agentId)
				initSucceeded = true
				break
			} else if _, hasError := jsonMap["error"]; hasError {
				uc.log.Errorf("MCP初始化失败，智能体ID: %s, 错误: %v", agentId, jsonMap["error"])
				return []string{}, nil
			}
		}
	}

	if !initSucceeded {
		uc.log.Errorf("未收到有效的MCP初始化响应，智能体ID: %s", agentId)
		return []string{}, nil
	}

	// 步骤2: 发送初始化完成通知
	uc.log.Infof("发送MCP初始化完成通知，智能体ID: %s", agentId)
	if err := client.SendText(kit.GetNotificationsInitializedJson()); err != nil {
		uc.log.Warnf("发送初始化完成通知失败: %v", err)
		return []string{}, nil
	}

	// 步骤3: 发送工具列表请求
	uc.log.Infof("发送MCP工具列表请求，智能体ID: %s", agentId)
	if err := client.SendText(kit.GetToolsListJson()); err != nil {
		uc.log.Warnf("发送工具列表请求失败: %v", err)
		return []string{}, nil
	}

	// 等待工具列表响应 (id=2)
	toolsResponses, err := client.ListenForResponse(func(response string) bool {
		jsonMap, err := kit.ParseJsonResponse(response)
		if err != nil {
			return false
		}
		if id, ok := jsonMap["id"].(float64); ok && int(id) == 2 {
			return true
		}
		return false
	})

	if err != nil {
		uc.log.Warnf("等待工具列表响应失败: %v", err)
		return []string{}, nil
	}

	// 处理工具列表响应
	for _, response := range toolsResponses {
		jsonMap, err := kit.ParseJsonResponse(response)
		if err != nil {
			continue
		}
		if id, ok := jsonMap["id"].(float64); ok && int(id) == 2 {
			// 检查是否有result字段
			if resultObj, hasResult := jsonMap["result"]; hasResult {
				if resultMap, ok := resultObj.(map[string]interface{}); ok {
					if toolsObj, hasTools := resultMap["tools"]; hasTools {
						if toolsList, ok := toolsObj.([]interface{}); ok {
							// 提取工具名称列表
							var result []string
							for _, tool := range toolsList {
								if toolMap, ok := tool.(map[string]interface{}); ok {
									if name, ok := toolMap["name"].(string); ok && name != "" {
										result = append(result, name)
									}
								}
							}
							uc.log.Infof("成功获取MCP工具列表，智能体ID: %s, 工具数量: %d", agentId, len(result))
							return result, nil
						}
					}
				}
			} else if _, hasError := jsonMap["error"]; hasError {
				uc.log.Errorf("获取工具列表失败，智能体ID: %s, 错误: %v", agentId, jsonMap["error"])
				return []string{}, nil
			}
		}
	}

	uc.log.Warnf("未找到有效的工具列表响应，智能体ID: %s", agentId)
	return []string{}, nil
}

// parseURI 解析URI
func parseURI(urlStr string) (*url.URL, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("路径格式不正确路径：%s，错误信息: %v", urlStr, err)
	}
	return parsedURL, nil
}

// getSecretKey 获取密钥
func (uc *AgentUsecase) getSecretKey(parsedURL *url.URL) string {
	query := parsedURL.RawQuery
	keyPrefix := "key="
	keyIndex := strings.Index(query, keyPrefix)
	if keyIndex == -1 {
		return ""
	}
	return query[keyIndex+len(keyPrefix):]
}

// getAgentMcpUrl 获取智能体mcp接入点url
func (uc *AgentUsecase) getAgentMcpUrl(parsedURL *url.URL) string {
	// 获取协议
	var wsScheme string
	if parsedURL.Scheme == "https" {
		wsScheme = "wss"
	} else {
		wsScheme = "ws"
	}

	// 获取主机和路径
	host := parsedURL.Host
	path := parsedURL.Path

	// 获取到最后一个/前的path
	lastSlashIndex := strings.LastIndex(path, "/")
	if lastSlashIndex != -1 {
		path = path[:lastSlashIndex]
	}

	return fmt.Sprintf("%s://%s%s", wsScheme, host, path)
}

// encryptToken 获取对智能体id加密的token
func (uc *AgentUsecase) encryptToken(agentId, key string) (string, error) {
	// 使用md5对智能体id进行加密
	md5 := kit.MD5HexDigest(agentId)

	// aes需要加密文本
	json := fmt.Sprintf(`{"agentId": "%s"}`, md5)

	// 加密后成token值
	return kit.AESEncrypt(key, json)
}

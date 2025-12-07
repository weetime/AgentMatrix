package biz

import (
	"context"
	"fmt"
	"net/url"
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

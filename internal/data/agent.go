package data

import (
	"context"
	"fmt"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/agent"
	"github.com/weetime/agent-matrix/internal/data/ent/agentchataudio"
	"github.com/weetime/agent-matrix/internal/data/ent/agentchathistory"
	"github.com/weetime/agent-matrix/internal/data/ent/agentpluginmapping"
	"github.com/weetime/agent-matrix/internal/data/ent/agenttemplate"
	"github.com/weetime/agent-matrix/internal/data/ent/device"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type agentRepo struct {
	data *Data
	log  *log.Helper
}

// NewAgentRepo 初始化 Agent Repo
func NewAgentRepo(data *Data, logger log.Logger) biz.AgentRepo {
	return &agentRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/agent")),
	}
}

// ListUserAgents 获取用户智能体列表
func (r *agentRepo) ListUserAgents(ctx context.Context, userId int64) ([]*biz.AgentDTO, error) {
	agents, err := r.data.db.Agent.Query().
		Where(agent.UserIDEQ(userId)).
		Order(ent.Desc(agent.FieldSort), ent.Desc(agent.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentDTO, len(agents))
	for i, a := range agents {
		// 获取设备数量
		deviceCount, err := r.GetDeviceCountByAgentID(ctx, a.ID)
		if err != nil {
			r.log.Warnf("Failed to get device count for agent %s: %v", a.ID, err)
			deviceCount = 0
		}

		result[i] = &biz.AgentDTO{
			ID:            a.ID,
			AgentName:     a.AgentName,
			SystemPrompt:  a.SystemPrompt,
			SummaryMemory: a.SummaryMemory,
			MemModelID:    a.MemModelID,
			DeviceCount:   int32(deviceCount),
			// 注意：模型名称需要通过 ModelService 查询，这里先留空
			// TTSModelName, TTSVoiceName, LLMModelName, VLLMModelName 需要关联查询
		}
	}

	return result, nil
}

// ListAllAgents 管理员列表（分页）
func (r *agentRepo) ListAllAgents(ctx context.Context, page *kit.PageRequest) ([]*biz.Agent, int, error) {
	query := r.data.db.Agent.Query()

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 应用分页
	applyPagination(query, page, agent.Columns)

	agents, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.Agent, len(agents))
	for i, a := range agents {
		result[i] = &biz.Agent{
			ID:              a.ID,
			UserID:          a.UserID,
			AgentCode:       a.AgentCode,
			AgentName:       a.AgentName,
			ASRModelID:      a.AsrModelID,
			VADModelID:      a.VadModelID,
			LLMModelID:      a.LlmModelID,
			VLLMModelID:     a.VllmModelID,
			TTSModelID:      a.TtsModelID,
			TTSVoiceID:      a.TtsVoiceID,
			MemModelID:      a.MemModelID,
			IntentModelID:   a.IntentModelID,
			ChatHistoryConf: int8(a.ChatHistoryConf),
			SystemPrompt:    a.SystemPrompt,
			SummaryMemory:   a.SummaryMemory,
			LangCode:        a.LangCode,
			Language:        a.Language,
			Sort:            int8(a.Sort),
			Creator:         a.Creator,
			CreatedAt:       a.CreatedAt,
			Updater:         a.Updater,
			UpdatedAt:       a.UpdatedAt,
		}
	}

	return result, total, nil
}

// GetAgentByID 获取详情（包含插件映射）
func (r *agentRepo) GetAgentByID(ctx context.Context, id string) (*biz.Agent, []*biz.AgentPluginMapping, error) {
	agentEntity, err := r.data.db.Agent.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	agent := &biz.Agent{
		ID:              agentEntity.ID,
		UserID:          agentEntity.UserID,
		AgentCode:       agentEntity.AgentCode,
		AgentName:       agentEntity.AgentName,
		ASRModelID:      agentEntity.AsrModelID,
		VADModelID:      agentEntity.VadModelID,
		LLMModelID:      agentEntity.LlmModelID,
		VLLMModelID:     agentEntity.VllmModelID,
		TTSModelID:      agentEntity.TtsModelID,
		TTSVoiceID:      agentEntity.TtsVoiceID,
		MemModelID:      agentEntity.MemModelID,
		IntentModelID:   agentEntity.IntentModelID,
		ChatHistoryConf: int8(agentEntity.ChatHistoryConf),
		SystemPrompt:    agentEntity.SystemPrompt,
		SummaryMemory:   agentEntity.SummaryMemory,
		LangCode:        agentEntity.LangCode,
		Language:        agentEntity.Language,
		Sort:            int8(agentEntity.Sort),
		Creator:         agentEntity.Creator,
		CreatedAt:       agentEntity.CreatedAt,
		Updater:         agentEntity.Updater,
		UpdatedAt:       agentEntity.UpdatedAt,
	}

	// 查询插件映射
	pluginMappingEntities, err := r.data.db.AgentPluginMapping.Query().
		Where(agentpluginmapping.AgentIDEQ(id)).
		All(ctx)
	if err != nil {
		// 如果查询失败，记录日志但不影响主流程
		r.log.Warnf("Failed to query plugin mappings for agent %s: %v", id, err)
		pluginMappings := []*biz.AgentPluginMapping{}
		return agent, pluginMappings, nil
	}

	// 转换为业务实体
	pluginMappings := make([]*biz.AgentPluginMapping, 0, len(pluginMappingEntities))
	for _, pm := range pluginMappingEntities {
		// 查询 provider_code（使用 ent 查询）
		providerCode := r.getProviderCode(ctx, pm.PluginID)

		pluginMappings = append(pluginMappings, &biz.AgentPluginMapping{
			ID:           pm.ID,
			AgentID:      pm.AgentID,
			PluginID:     pm.PluginID,
			ParamInfo:    pm.ParamInfo,
			ProviderCode: providerCode,
		})
	}

	return agent, pluginMappings, nil
}

// CreateAgent 创建（自动生成 ID 和 agentCode）
func (r *agentRepo) CreateAgent(ctx context.Context, agent *biz.Agent) (*biz.Agent, error) {
	// ID 和 agentCode 应该在 biz 层生成，这里直接使用
	create := r.data.db.Agent.Create().
		SetID(agent.ID).
		SetNillableUserID(&agent.UserID).
		SetNillableAgentCode(&agent.AgentCode).
		SetNillableAgentName(&agent.AgentName).
		SetNillableAsrModelID(&agent.ASRModelID).
		SetNillableVadModelID(&agent.VADModelID).
		SetNillableLlmModelID(&agent.LLMModelID).
		SetNillableVllmModelID(&agent.VLLMModelID).
		SetNillableTtsModelID(&agent.TTSModelID).
		SetNillableTtsVoiceID(&agent.TTSVoiceID).
		SetNillableMemModelID(&agent.MemModelID).
		SetNillableIntentModelID(&agent.IntentModelID).
		SetChatHistoryConf(int32(agent.ChatHistoryConf)).
		SetNillableSystemPrompt(&agent.SystemPrompt).
		SetNillableSummaryMemory(&agent.SummaryMemory).
		SetNillableLangCode(&agent.LangCode).
		SetNillableLanguage(&agent.Language).
		SetSort(int32(agent.Sort))

	if agent.Creator > 0 {
		create.SetCreator(agent.Creator)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.Agent{
		ID:              entity.ID,
		UserID:          entity.UserID,
		AgentCode:       entity.AgentCode,
		AgentName:       entity.AgentName,
		ASRModelID:      entity.AsrModelID,
		VADModelID:      entity.VadModelID,
		LLMModelID:      entity.LlmModelID,
		VLLMModelID:     entity.VllmModelID,
		TTSModelID:      entity.TtsModelID,
		TTSVoiceID:      entity.TtsVoiceID,
		MemModelID:      entity.MemModelID,
		IntentModelID:   entity.IntentModelID,
		ChatHistoryConf: int8(entity.ChatHistoryConf),
		SystemPrompt:    entity.SystemPrompt,
		SummaryMemory:   entity.SummaryMemory,
		LangCode:        entity.LangCode,
		Language:        entity.Language,
		Sort:            int8(entity.Sort),
		Creator:         entity.Creator,
		CreatedAt:       entity.CreatedAt,
		Updater:         entity.Updater,
		UpdatedAt:       entity.UpdatedAt,
	}, nil
}

// UpdateAgent 更新
func (r *agentRepo) UpdateAgent(ctx context.Context, agent *biz.Agent) error {
	update := r.data.db.Agent.UpdateOneID(agent.ID)

	if agent.AgentCode != "" {
		update.SetAgentCode(agent.AgentCode)
	}
	if agent.AgentName != "" {
		update.SetAgentName(agent.AgentName)
	}
	if agent.ASRModelID != "" {
		update.SetAsrModelID(agent.ASRModelID)
	}
	if agent.VADModelID != "" {
		update.SetVadModelID(agent.VADModelID)
	}
	if agent.LLMModelID != "" {
		update.SetLlmModelID(agent.LLMModelID)
	}
	if agent.VLLMModelID != "" {
		update.SetVllmModelID(agent.VLLMModelID)
	}
	if agent.TTSModelID != "" {
		update.SetTtsModelID(agent.TTSModelID)
	}
	if agent.TTSVoiceID != "" {
		update.SetTtsVoiceID(agent.TTSVoiceID)
	}
	if agent.MemModelID != "" {
		update.SetMemModelID(agent.MemModelID)
	}
	if agent.IntentModelID != "" {
		update.SetIntentModelID(agent.IntentModelID)
	}
	if agent.SystemPrompt != "" {
		update.SetSystemPrompt(agent.SystemPrompt)
	}
	if agent.SummaryMemory != "" {
		update.SetSummaryMemory(agent.SummaryMemory)
	}
	if agent.LangCode != "" {
		update.SetLangCode(agent.LangCode)
	}
	if agent.Language != "" {
		update.SetLanguage(agent.Language)
	}
	if agent.Updater > 0 {
		update.SetUpdater(agent.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteAgent 删除
func (r *agentRepo) DeleteAgent(ctx context.Context, id string) error {
	_, err := r.data.db.Agent.Delete().Where(agent.IDEQ(id)).Exec(ctx)
	return err
}

// GetAgentTemplateList 模板列表
func (r *agentRepo) GetAgentTemplateList(ctx context.Context) ([]*biz.AgentTemplate, error) {
	templates, err := r.data.db.AgentTemplate.Query().
		Order(ent.Asc(agenttemplate.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentTemplate, len(templates))
	for i, t := range templates {
		result[i] = &biz.AgentTemplate{
			ID:              t.ID,
			AgentCode:       t.AgentCode,
			AgentName:       t.AgentName,
			ASRModelID:      t.AsrModelID,
			VADModelID:      t.VadModelID,
			LLMModelID:      t.LlmModelID,
			VLLMModelID:     t.VllmModelID,
			TTSModelID:      t.TtsModelID,
			TTSVoiceID:      t.TtsVoiceID,
			MemModelID:      t.MemModelID,
			IntentModelID:   t.IntentModelID,
			ChatHistoryConf: int8(t.ChatHistoryConf),
			SystemPrompt:    t.SystemPrompt,
			SummaryMemory:   t.SummaryMemory,
			LangCode:        t.LangCode,
			Language:        t.Language,
			Sort:            int8(t.Sort),
		}
	}

	return result, nil
}

// GetSessionsByAgentID 会话列表（分组统计）
func (r *agentRepo) GetSessionsByAgentID(ctx context.Context, agentId string, page *kit.PageRequest) ([]*biz.AgentChatSession, int, error) {
	// 查询所有聊天记录
	histories, err := r.data.db.AgentChatHistory.Query().
		Where(agentchathistory.AgentIDEQ(agentId)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 手动分组统计
	sessionMap := make(map[string]*biz.AgentChatSession)
	for _, h := range histories {
		if h.SessionID == "" {
			continue
		}
		sessionID := h.SessionID
		if s, exists := sessionMap[sessionID]; exists {
			s.ChatCount++
			// 更新最早创建时间（会话时间）
			if h.CreatedAt.Before(s.CreatedAt) {
				s.CreatedAt = h.CreatedAt
			}
		} else {
			sessionMap[sessionID] = &biz.AgentChatSession{
				SessionID: sessionID,
				CreatedAt: h.CreatedAt,
				ChatCount: 1,
			}
		}
	}

	// 转换为切片并排序
	sessions := make([]*biz.AgentChatSession, 0, len(sessionMap))
	for _, s := range sessionMap {
		sessions = append(sessions, s)
	}

	// 按创建时间降序排序（使用简单的冒泡排序）
	for i := 0; i < len(sessions)-1; i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[i].CreatedAt.Before(sessions[j].CreatedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	total := len(sessions)

	// 应用分页
	if page != nil {
		pageSize := int(page.GetPageSize())
		pageNo, _ := page.GetPageNo()
		if pageNo > 0 && pageSize > 0 {
			offset := (pageNo - 1) * pageSize
			if offset < len(sessions) {
				end := offset + pageSize
				if end > len(sessions) {
					end = len(sessions)
				}
				sessions = sessions[offset:end]
			} else {
				sessions = []*biz.AgentChatSession{}
			}
		}
	}

	return sessions, total, nil
}

// GetChatHistoryBySessionID 聊天记录
func (r *agentRepo) GetChatHistoryBySessionID(ctx context.Context, agentId, sessionId string) ([]*biz.AgentChatHistory, error) {
	histories, err := r.data.db.AgentChatHistory.Query().
		Where(
			agentchathistory.AgentIDEQ(agentId),
			agentchathistory.SessionIDEQ(sessionId),
		).
		Order(ent.Asc(agentchathistory.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentChatHistory, len(histories))
	for i, h := range histories {
		var macAddress, agentID, sessionID, content, audioID *string
		if h.MACAddress != "" {
			macAddress = &h.MACAddress
		}
		if h.AgentID != "" {
			agentID = &h.AgentID
		}
		if h.SessionID != "" {
			sessionID = &h.SessionID
		}
		if h.Content != "" {
			content = &h.Content
		}
		if h.AudioID != "" {
			audioID = &h.AudioID
		}
		result[i] = &biz.AgentChatHistory{
			ID:         h.ID,
			MacAddress: macAddress,
			AgentID:    agentID,
			SessionID:  sessionID,
			ChatType:   int8(h.ChatType),
			Content:    content,
			AudioID:    audioID,
			CreatedAt:  h.CreatedAt,
			UpdatedAt:  h.UpdatedAt,
		}
	}

	return result, nil
}

// GetRecentFiftyUserChats 最近50条用户聊天
func (r *agentRepo) GetRecentFiftyUserChats(ctx context.Context, agentId string) ([]*biz.AgentChatHistoryUserVO, error) {
	histories, err := r.data.db.AgentChatHistory.Query().
		Where(
			agentchathistory.AgentIDEQ(agentId),
			agentchathistory.ChatTypeEQ(1), // 1-用户
		).
		Order(ent.Desc(agentchathistory.FieldCreatedAt)).
		Limit(50).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentChatHistoryUserVO, len(histories))
	for i, h := range histories {
		audioID := ""
		if h.AudioID != "" {
			audioID = h.AudioID
		}
		result[i] = &biz.AgentChatHistoryUserVO{
			Content: h.Content,
			AudioID: audioID,
		}
	}

	return result, nil
}

// GetContentByAudioID 根据音频ID获取内容
func (r *agentRepo) GetContentByAudioID(ctx context.Context, audioId string) (string, error) {
	history, err := r.data.db.AgentChatHistory.Query().
		Where(agentchathistory.AudioIDEQ(audioId)).
		Only(ctx)
	if err != nil {
		return "", err
	}

	if history.Content == "" {
		return "", nil
	}
	return history.Content, nil
}

// GetAudioByID 获取音频数据
func (r *agentRepo) GetAudioByID(ctx context.Context, audioId string) ([]byte, error) {
	audio, err := r.data.db.AgentChatAudio.Get(ctx, audioId)
	if err != nil {
		return nil, err
	}

	return audio.Audio, nil
}

// SaveAudio 保存音频
func (r *agentRepo) SaveAudio(ctx context.Context, audioId string, audioData []byte) error {
	_, err := r.data.db.AgentChatAudio.Create().
		SetID(audioId).
		SetAudio(audioData).
		Save(ctx)
	return err
}

// DeleteChatHistoryByAgentID 删除智能体的聊天记录
func (r *agentRepo) DeleteChatHistoryByAgentID(ctx context.Context, agentId string) error {
	_, err := r.data.db.AgentChatHistory.Delete().
		Where(agentchathistory.AgentIDEQ(agentId)).
		Exec(ctx)
	return err
}

// DeleteAudioByAgentID 删除智能体的音频数据
func (r *agentRepo) DeleteAudioByAgentID(ctx context.Context, agentId string) error {
	// 先查询该智能体的所有 audio_id
	histories, err := r.data.db.AgentChatHistory.Query().
		Where(agentchathistory.AgentIDEQ(agentId)).
		Select(agentchathistory.FieldAudioID).
		All(ctx)
	if err != nil {
		return err
	}

	audioIDs := make([]string, 0)
	for _, h := range histories {
		if h.AudioID != "" {
			audioIDs = append(audioIDs, h.AudioID)
		}
	}

	if len(audioIDs) > 0 {
		_, err = r.data.db.AgentChatAudio.Delete().
			Where(agentchataudio.IDIn(audioIDs...)).
			Exec(ctx)
		return err
	}

	return nil
}

// GetDeviceCountByAgentID 获取智能体的设备数量
func (r *agentRepo) GetDeviceCountByAgentID(ctx context.Context, agentId string) (int, error) {
	count, err := r.data.db.Device.Query().
		Where(device.AgentIDEQ(agentId)).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetDefaultAgentByMacAddress 根据 MAC 地址获取默认智能体
func (r *agentRepo) GetDefaultAgentByMacAddress(ctx context.Context, macAddress string) (*biz.Agent, error) {
	// 需要 Device schema 来关联查询
	// TODO: 实现通过设备 MAC 地址查询智能体
	return nil, fmt.Errorf("not implemented: need Device schema")
}

// getProviderCode 通过 ent 查询 provider_code
func (r *agentRepo) getProviderCode(ctx context.Context, pluginID string) string {
	// 使用 ent 查询 provider_code
	provider, err := r.data.db.ModelProvider.Get(ctx, pluginID)
	if err != nil {
		// 如果查询失败（记录不存在或其他错误），返回空字符串
		if !ent.IsNotFound(err) {
			r.log.Debugf("Failed to query provider_code for plugin %s: %v", pluginID, err)
		}
		return ""
	}
	return provider.ProviderCode
}

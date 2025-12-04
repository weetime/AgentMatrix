package service

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type AgentService struct {
	uc *biz.AgentUsecase
	pb.UnimplementedAgentServiceServer
}

func NewAgentService(uc *biz.AgentUsecase) *AgentService {
	return &AgentService{
		uc: uc,
	}
}

// ListUserAgents 获取用户智能体列表
func (s *AgentService) ListUserAgents(ctx context.Context, req *pb.Empty) (*pb.Response, error) {
	// TODO: 从 context 获取用户ID（需要实现认证中间件）
	userId := int64(1) // 临时值，后续需要从 context 获取

	agents, err := s.uc.ListUserAgents(ctx, userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(agents))
	for _, agent := range agents {
		dtoList = append(dtoList, map[string]interface{}{
			"id":              agent.ID,
			"agentName":       agent.AgentName,
			"ttsModelName":    agent.TTSModelName,
			"ttsVoiceName":    agent.TTSVoiceName,
			"llmModelName":    agent.LLMModelName,
			"vllmModelName":   agent.VLLMModelName,
			"memModelId":      agent.MemModelID,
			"systemPrompt":    agent.SystemPrompt,
			"summaryMemory":   agent.SummaryMemory,
			"lastConnectedAt": agent.LastConnectedAt,
			"deviceCount":     agent.DeviceCount,
		})
	}

	data := map[string]interface{}{
		"list": dtoList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// ListAllAgents 智能体列表（管理员）
func (s *AgentService) ListAllAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.Response, error) {
	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("created_at")

	agents, total, err := s.uc.ListAllAgents(ctx, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(agents))
	for _, agent := range agents {
		dtoList = append(dtoList, map[string]interface{}{
			"id":              agent.ID,
			"userId":          agent.UserID,
			"agentCode":       agent.AgentCode,
			"agentName":       agent.AgentName,
			"asrModelId":      agent.ASRModelID,
			"vadModelId":      agent.VADModelID,
			"llmModelId":      agent.LLMModelID,
			"vllmModelId":     agent.VLLMModelID,
			"ttsModelId":      agent.TTSModelID,
			"ttsVoiceId":      agent.TTSVoiceID,
			"memModelId":      agent.MemModelID,
			"intentModelId":   agent.IntentModelID,
			"chatHistoryConf": agent.ChatHistoryConf,
			"systemPrompt":    agent.SystemPrompt,
			"summaryMemory":   agent.SummaryMemory,
			"langCode":        agent.LangCode,
			"language":        agent.Language,
			"sort":            agent.Sort,
			"creator":         agent.Creator,
			"createdAt":       agent.CreatedAt.Format(time.RFC3339),
			"updater":         agent.Updater,
			"updatedAt":       agent.UpdatedAt.Format(time.RFC3339),
		})
	}

	data := map[string]interface{}{
		"total": int32(total),
		"list":  dtoList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetAgentById 获取智能体详情
func (s *AgentService) GetAgentById(ctx context.Context, req *pb.GetAgentByIdRequest) (*pb.Response, error) {
	agent, pluginMappings, err := s.uc.GetAgentByID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	// 转换插件映射
	functions := make([]interface{}, 0, len(pluginMappings))
	for _, pm := range pluginMappings {
		functions = append(functions, map[string]interface{}{
			"id":           pm.ID,
			"agentId":      pm.AgentID,
			"pluginId":     pm.PluginID,
			"paramInfo":    pm.ParamInfo,
			"providerCode": pm.ProviderCode,
		})
	}

	data := map[string]interface{}{
		"id":              agent.ID,
		"userId":          agent.UserID,
		"agentCode":       agent.AgentCode,
		"agentName":       agent.AgentName,
		"asrModelId":      agent.ASRModelID,
		"vadModelId":      agent.VADModelID,
		"llmModelId":      agent.LLMModelID,
		"vllmModelId":     agent.VLLMModelID,
		"ttsModelId":      agent.TTSModelID,
		"ttsVoiceId":      agent.TTSVoiceID,
		"memModelId":      agent.MemModelID,
		"intentModelId":   agent.IntentModelID,
		"chatHistoryConf": agent.ChatHistoryConf,
		"systemPrompt":    agent.SystemPrompt,
		"summaryMemory":   agent.SummaryMemory,
		"langCode":        agent.LangCode,
		"language":        agent.Language,
		"sort":            agent.Sort,
		"creator":         agent.Creator,
		"createdAt":       agent.CreatedAt.Format(time.RFC3339),
		"updater":         agent.Updater,
		"updatedAt":       agent.UpdatedAt.Format(time.RFC3339),
		"functions":       functions,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// CreateAgent 创建智能体
func (s *AgentService) CreateAgent(ctx context.Context, req *pb.AgentCreateRequest) (*pb.Response, error) {
	if req == nil || req.GetAgentName() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体名称不能为空",
		}, nil
	}

	// TODO: 从 context 获取用户ID
	userId := int64(1)

	agent, err := s.uc.CreateAgent(ctx, req.GetAgentName(), userId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	data := map[string]interface{}{
		"id": agent.ID,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// UpdateAgentMemoryByMacAddress 根据设备更新智能体记忆
func (s *AgentService) UpdateAgentMemoryByMacAddress(ctx context.Context, req *pb.UpdateAgentMemoryByMacAddressRequest) (*pb.Response, error) {
	if req == nil || req.GetMacAddress() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "MAC地址不能为空",
		}, nil
	}

	err := s.uc.UpdateAgentMemoryByMacAddress(ctx, req.GetMacAddress(), req.GetSummaryMemory())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// UpdateAgent 更新智能体
func (s *AgentService) UpdateAgent(ctx context.Context, req *pb.AgentUpdateRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体ID不能为空",
		}, nil
	}

	agent := &biz.Agent{
		ID: req.GetId(),
	}

	if req.AgentCode != nil {
		agent.AgentCode = req.AgentCode.GetValue()
	}
	if req.AgentName != nil {
		agent.AgentName = req.AgentName.GetValue()
	}
	if req.AsrModelId != nil {
		agent.ASRModelID = req.AsrModelId.GetValue()
	}
	if req.VadModelId != nil {
		agent.VADModelID = req.VadModelId.GetValue()
	}
	if req.LlmModelId != nil {
		agent.LLMModelID = req.LlmModelId.GetValue()
	}
	if req.VllmModelId != nil {
		agent.VLLMModelID = req.VllmModelId.GetValue()
	}
	if req.TtsModelId != nil {
		agent.TTSModelID = req.TtsModelId.GetValue()
	}
	if req.TtsVoiceId != nil {
		agent.TTSVoiceID = req.TtsVoiceId.GetValue()
	}
	if req.MemModelId != nil {
		agent.MemModelID = req.MemModelId.GetValue()
	}
	if req.IntentModelId != nil {
		agent.IntentModelID = req.IntentModelId.GetValue()
	}
	if req.SystemPrompt != nil {
		agent.SystemPrompt = req.SystemPrompt.GetValue()
	}
	if req.SummaryMemory != nil {
		agent.SummaryMemory = req.SummaryMemory.GetValue()
	}
	if req.LangCode != nil {
		agent.LangCode = req.LangCode.GetValue()
	}
	if req.Language != nil {
		agent.Language = req.Language.GetValue()
	}
	if req.Sort != nil {
		agent.Sort = int8(req.Sort.GetValue())
	}

	// TODO: 从 context 获取用户ID
	agent.Updater = int64(1)

	err := s.uc.UpdateAgent(ctx, agent)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// DeleteAgent 删除智能体（级联删除）
func (s *AgentService) DeleteAgent(ctx context.Context, req *pb.GetAgentByIdRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体ID不能为空",
		}, nil
	}

	err := s.uc.DeleteAgent(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetAgentTemplates 获取智能体模板列表
func (s *AgentService) GetAgentTemplates(ctx context.Context, req *pb.Empty) (*pb.Response, error) {
	templates, err := s.uc.GetAgentTemplateList(ctx)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	templateList := make([]interface{}, 0, len(templates))
	for _, t := range templates {
		templateList = append(templateList, map[string]interface{}{
			"id":              t.ID,
			"agentCode":       t.AgentCode,
			"agentName":       t.AgentName,
			"asrModelId":      t.ASRModelID,
			"vadModelId":      t.VADModelID,
			"llmModelId":      t.LLMModelID,
			"vllmModelId":     t.VLLMModelID,
			"ttsModelId":      t.TTSModelID,
			"ttsVoiceId":      t.TTSVoiceID,
			"memModelId":      t.MemModelID,
			"intentModelId":   t.IntentModelID,
			"chatHistoryConf": t.ChatHistoryConf,
			"systemPrompt":    t.SystemPrompt,
			"summaryMemory":   t.SummaryMemory,
			"langCode":        t.LangCode,
			"language":        t.Language,
			"sort":            t.Sort,
		})
	}

	data := map[string]interface{}{
		"list": templateList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetAgentSessions 获取智能体会话列表
func (s *AgentService) GetAgentSessions(ctx context.Context, req *pb.GetAgentSessionsRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体ID不能为空",
		}, nil
	}

	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))

	sessions, total, err := s.uc.GetAgentSessions(ctx, req.GetId(), page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	sessionList := make([]interface{}, 0, len(sessions))
	for _, session := range sessions {
		sessionList = append(sessionList, map[string]interface{}{
			"sessionId": session.SessionID,
			"createdAt": session.CreatedAt.Format(time.RFC3339),
			"chatCount": session.ChatCount,
		})
	}

	data := map[string]interface{}{
		"total": int32(total),
		"list":  sessionList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetAgentChatHistory 获取智能体聊天记录
func (s *AgentService) GetAgentChatHistory(ctx context.Context, req *pb.GetAgentChatHistoryRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" || req.GetSessionId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体ID和会话ID不能为空",
		}, nil
	}

	histories, err := s.uc.GetAgentChatHistory(ctx, req.GetId(), req.GetSessionId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	historyList := make([]interface{}, 0, len(histories))
	for _, h := range histories {
		m := map[string]interface{}{
			"createdAt": h.CreatedAt.Format(time.RFC3339),
			"chatType":  h.ChatType,
		}
		if h.Content != nil {
			m["content"] = *h.Content
		}
		if h.AudioID != nil {
			m["audioId"] = *h.AudioID
		}
		if h.MacAddress != nil {
			m["macAddress"] = *h.MacAddress
		}
		historyList = append(historyList, m)
	}

	data := map[string]interface{}{
		"list": historyList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetRecentFiftyUserChats 获取智能体最近50条聊天记录
func (s *AgentService) GetRecentFiftyUserChats(ctx context.Context, req *pb.GetAgentByIdRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "智能体ID不能为空",
		}, nil
	}

	chats, err := s.uc.GetRecentFiftyUserChats(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	chatList := make([]interface{}, 0, len(chats))
	for _, chat := range chats {
		chatList = append(chatList, map[string]interface{}{
			"content": chat.Content,
			"audioId": chat.AudioID,
		})
	}

	data := map[string]interface{}{
		"list": chatList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetContentByAudioId 获取音频内容
func (s *AgentService) GetContentByAudioId(ctx context.Context, req *pb.GetAgentByIdRequest) (*pb.Response, error) {
	if req == nil || req.GetId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "音频ID不能为空",
		}, nil
	}

	content, err := s.uc.GetContentByAudioID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	data := map[string]interface{}{
		"content": content,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetAudioDownloadID 获取音频下载ID
func (s *AgentService) GetAudioDownloadID(ctx context.Context, req *pb.GetAudioDownloadIDRequest) (*pb.Response, error) {
	if req == nil || req.GetAudioId() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "音频ID不能为空",
		}, nil
	}

	uuidStr, err := s.uc.GetAudioDownloadID(ctx, req.GetAudioId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	data := map[string]interface{}{
		"uuid": uuidStr,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// PlayAudio 播放音频
func (s *AgentService) PlayAudio(ctx context.Context, req *pb.PlayAudioRequest) (*pb.Response, error) {
	if req == nil || req.GetUuid() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "UUID不能为空",
		}, nil
	}

	audioData, err := s.uc.PlayAudio(ctx, req.GetUuid())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	// 返回音频数据的 base64 编码
	data := map[string]interface{}{
		"audio": fmt.Sprintf("%x", audioData), // 使用十六进制编码
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/constant"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type ModelService struct {
	uc            *biz.ModelUsecase
	providerUC    *biz.ModelProviderUsecase
	ttsVoiceUC    *biz.TtsVoiceUsecase
	voiceCloneUC  *biz.VoiceCloneUsecase
	configService *ConfigService // 用于刷新配置缓存
	pb.UnimplementedModelServiceServer
}

func NewModelService(uc *biz.ModelUsecase, providerUC *biz.ModelProviderUsecase, ttsVoiceUC *biz.TtsVoiceUsecase, voiceCloneUC *biz.VoiceCloneUsecase, configService *ConfigService) *ModelService {
	return &ModelService{
		uc:            uc,
		providerUC:    providerUC,
		ttsVoiceUC:    ttsVoiceUC,
		voiceCloneUC:  voiceCloneUC,
		configService: configService,
	}
}

// GetModelNames 获取所有模型名称
func (s *ModelService) GetModelNames(ctx context.Context, req *pb.GetModelNamesRequest) (*pb.Response, error) {
	var modelName *string
	if req.ModelName != nil && req.ModelName.GetValue() != "" {
		name := req.ModelName.GetValue()
		modelName = &name
	}

	list, err := s.uc.GetModelCodeList(ctx, req.GetModelType(), modelName)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(list))
	for _, item := range list {
		dtoList = append(dtoList, map[string]interface{}{
			"id":        item.ID, // ID已经是string类型
			"modelName": item.ModelName,
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

// GetLlmModelNames 获取LLM模型信息
func (s *ModelService) GetLlmModelNames(ctx context.Context, req *pb.GetLlmModelNamesRequest) (*pb.Response, error) {
	var modelName *string
	if req.ModelName != nil && req.ModelName.GetValue() != "" {
		name := req.ModelName.GetValue()
		modelName = &name
	}

	list, err := s.uc.GetLlmModelCodeList(ctx, modelName)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表，从configJson中提取type字段
	dtoList := make([]interface{}, 0, len(list))
	for _, item := range list {
		dto := map[string]interface{}{
			"id":        item.ID,
			"modelName": item.ModelName,
			"type":      "",
		}

		// 从configJson中提取type字段
		if item.ConfigJSON != "" {
			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(item.ConfigJSON), &configMap); err == nil {
				if typeValue, ok := configMap["type"].(string); ok {
					dto["type"] = typeValue
				}
			}
		}

		dtoList = append(dtoList, dto)
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

// GetModelProviderList 获取模型供应器列表
func (s *ModelService) GetModelProviderList(ctx context.Context, req *pb.GetModelProviderListRequest) (*pb.Response, error) {
	list, err := s.uc.GetModelProviderList(ctx, req.GetModelType())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(list))
	for _, item := range list {
		dtoList = append(dtoList, map[string]interface{}{
			"id":           item.ID,
			"modelType":    item.ModelType,
			"providerCode": item.ProviderCode,
			"name":         item.Name,
			"fields":       item.Fields,
			"sort":         item.Sort,
			"updater":      fmt.Sprintf("%d", item.Updater),
			"updateDate":   item.UpdateDate.Format(time.RFC3339),
			"creator":      fmt.Sprintf("%d", item.Creator),
			"createDate":   item.CreateDate.Format(time.RFC3339),
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

// PageModelConfig 获取模型配置列表
func (s *ModelService) PageModelConfig(ctx context.Context, req *pb.PageModelConfigRequest) (*pb.Response, error) {
	// 解析过滤条件
	params := &biz.ListModelConfigParams{
		ModelType: req.GetModelType(),
	}
	if req.ModelName != nil && req.ModelName.GetValue() != "" {
		name := req.ModelName.GetValue()
		params.ModelName = &name
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1 // 默认第1页
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = 10 // 默认10
	}
	page.SetPageNo(int(pageNo)) // 前端传的page从1开始，直接使用
	page.SetPageSize(int(pageSize))

	// 查询列表和总数
	list, total, err := s.uc.GetModelConfigList(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为DTO列表
	dtoList := make([]interface{}, 0, len(list))
	for _, item := range list {
		dto := s.modelConfigToDTO(item)
		dtoList = append(dtoList, dto)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": total,
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

// AddModelConfig 新增模型配置
func (s *ModelService) AddModelConfig(ctx context.Context, req *pb.AddModelConfigRequest) (*pb.Response, error) {
	if req.Body == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "配置数据不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	// 构建业务对象
	configID := req.Body.GetId()
	config := &biz.ModelConfig{
		ID:         configID,
		ModelCode:  req.Body.ModelCode,
		ModelName:  req.Body.ModelName,
		IsDefault:  req.Body.IsDefault == 1,
		IsEnabled:  req.Body.IsEnabled == 1,
		DocLink:    req.Body.DocLink,
		Remark:     req.Body.Remark,
		Sort:       req.Body.Sort,
		Creator:    userID,
		Updater:    userID,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}

	// 转换configJson
	if req.Body.ConfigJson != nil {
		configJSONBytes, err := req.Body.ConfigJson.MarshalJSON()
		if err == nil {
			config.ConfigJSON = string(configJSONBytes)
		}
	}

	// 调用业务逻辑
	created, err := s.uc.AddModelConfig(ctx, req.GetModelType(), req.GetProvideCode(), config)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 刷新配置缓存
	if s.configService != nil {
		s.configService.GetServerConfig(ctx, &pb.GetServerConfigRequest{IsCache: false})
	}

	// 返回DTO
	dto := s.modelConfigToDTO(created)
	dataStruct, err := structpb.NewStruct(dto)
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

// EditModelConfig 编辑模型配置
func (s *ModelService) EditModelConfig(ctx context.Context, req *pb.EditModelConfigRequest) (*pb.Response, error) {
	if req.Body == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "配置数据不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	// 构建业务对象
	config := &biz.ModelConfig{
		ModelCode: req.Body.ModelCode,
		ModelName: req.Body.ModelName,
		IsDefault: req.Body.IsDefault == 1,
		IsEnabled: req.Body.IsEnabled == 1,
		DocLink:   req.Body.DocLink,
		Remark:    req.Body.Remark,
		Sort:      req.Body.Sort,
		Updater:   userID,
	}

	// 转换configJson
	if req.Body.ConfigJson != nil {
		configJSONBytes, err := req.Body.ConfigJson.MarshalJSON()
		if err == nil {
			config.ConfigJSON = string(configJSONBytes)
		}
	}

	// 调用业务逻辑
	updated, err := s.uc.EditModelConfig(ctx, req.GetModelType(), req.GetProvideCode(), req.GetId(), config)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 刷新配置缓存
	if s.configService != nil {
		s.configService.GetServerConfig(ctx, &pb.GetServerConfigRequest{IsCache: false})
	}

	// 返回DTO
	dto := s.modelConfigToDTO(updated)
	dataStruct, err := structpb.NewStruct(dto)
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

// DeleteModelConfig 删除模型配置
func (s *ModelService) DeleteModelConfig(ctx context.Context, req *pb.DeleteModelConfigRequest) (*pb.Response, error) {
	err := s.uc.DeleteModelConfig(ctx, req.GetId())
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

// GetModelConfig 获取模型配置
func (s *ModelService) GetModelConfig(ctx context.Context, req *pb.GetModelConfigRequest) (*pb.Response, error) {
	config, err := s.uc.GetModelConfigByID(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 404,
			Msg:  err.Error(),
		}, nil
	}

	// 返回DTO
	dto := s.modelConfigToDTO(config)
	dataStruct, err := structpb.NewStruct(dto)
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

// EnableModelConfig 启用/关闭模型
func (s *ModelService) EnableModelConfig(ctx context.Context, req *pb.EnableModelConfigRequest) (*pb.Response, error) {
	status := req.GetStatus() == 1
	err := s.uc.EnableModelConfig(ctx, req.GetId(), status)
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

// SetDefaultModel 设置默认模型
func (s *ModelService) SetDefaultModel(ctx context.Context, req *pb.SetDefaultModelRequest) (*pb.Response, error) {
	_, err := s.uc.SetDefaultModel(ctx, req.GetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 刷新配置缓存
	if s.configService != nil {
		s.configService.GetServerConfig(ctx, &pb.GetServerConfigRequest{IsCache: false})
	}

	// TODO: 更新模板表中对应的模型ID（需要实现AgentTemplateService）
	// agentTemplateService.updateDefaultTemplateModelId(config.ModelType, config.ID)

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetModelVoices 获取模型音色
func (s *ModelService) GetModelVoices(ctx context.Context, req *pb.GetModelVoicesRequest) (*pb.Response, error) {
	modelId := req.GetModelId()
	if modelId == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "modelId不能为空",
		}, nil
	}

	// 解析可选的 voice_name 参数
	var voiceName *string
	if req.VoiceName != nil && req.VoiceName.GetValue() != "" {
		name := req.VoiceName.GetValue()
		voiceName = &name
	}

	// 1. 查询普通音色（根据 tts_model_id 和 voice_name 过滤）
	params := &biz.ListTtsVoiceParams{
		TtsModelID: modelId,
		Name:       voiceName,
	}
	// 创建一个不分页的查询（设置一个很大的 limit）
	page := &kit.PageRequest{}
	page.SetPageNo(1)
	page.SetPageSize(10000) // 设置一个很大的值，实际返回所有结果
	page.SetSortAsc()
	page.SetSortField("sort")

	ttsVoices, err := s.ttsVoiceUC.ListTtsVoice(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "查询音色失败: " + err.Error(),
		}, nil
	}

	// 2. 转换为 VoiceDTO 格式
	voiceDTOs := make([]*biz.VoiceDTO, 0, len(ttsVoices))
	for _, voice := range ttsVoices {
		voiceDTOs = append(voiceDTOs, &biz.VoiceDTO{
			ID:        voice.ID,
			Name:      voice.Name,
			VoiceDemo: voice.VoiceDemo,
		})
	}

	// 3. 获取当前登录用户ID
	userID, err := middleware.GetUserIdFromContext(ctx)
	if err == nil && userID > 0 {
		// 4. 查询用户训练成功的克隆音色
		cloneVoices, err := s.voiceCloneUC.GetTrainSuccess(ctx, modelId, userID)
		if err != nil {
			// 查询克隆音色失败不影响主流程，记录日志但继续执行
			// 这里可以选择记录日志或忽略错误
		} else if len(cloneVoices) > 0 {
			// 5. 将克隆音色添加到列表前面，并在名称前加上前缀
			prefixedClones := make([]*biz.VoiceDTO, len(cloneVoices))
			for i, clone := range cloneVoices {
				prefixedClones[i] = &biz.VoiceDTO{
					ID:        clone.ID,
					Name:      constant.VoiceClonePrefix + clone.Name,
					VoiceDemo: clone.VoiceDemo,
				}
			}
			// 将克隆音色添加到列表前面
			voiceDTOs = append(prefixedClones, voiceDTOs...)

			// 6. 将克隆音色名称缓存到 Redis（永久过期）
			// 注意：需要从 Data 中获取 Redis 客户端，这里暂时跳过
			// 如果 ModelService 需要访问 Redis，可以在结构体中添加 RedisClient 字段
			// 为了简化，这里先不实现 Redis 缓存，后续可以根据需要添加
		}
	}

	// 7. 转换为接口列表
	dtoList := make([]interface{}, len(voiceDTOs))
	for i, dto := range voiceDTOs {
		dtoMap := map[string]interface{}{
			"id":   dto.ID,
			"name": dto.Name,
		}
		if dto.VoiceDemo != "" {
			dtoMap["voiceDemo"] = dto.VoiceDemo
		}
		dtoList[i] = dtoMap
	}

	// 8. 构建响应数据
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

// modelConfigToDTO 将业务对象转换为DTO
func (s *ModelService) modelConfigToDTO(config *biz.ModelConfig) map[string]interface{} {
	dto := map[string]interface{}{
		"id":        config.ID,
		"modelType": config.ModelType,
		"modelCode": config.ModelCode,
		"modelName": config.ModelName,
		"isDefault": 0,
		"isEnabled": 0,
		"docLink":   config.DocLink,
		"remark":    config.Remark,
		"sort":      config.Sort,
	}

	if config.IsDefault {
		dto["isDefault"] = 1
	}
	if config.IsEnabled {
		dto["isEnabled"] = 1
	}

	// 处理configJson
	if config.ConfigJSON != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(config.ConfigJSON), &configMap); err == nil {
			// 对敏感数据进行掩码处理
			maskedMap, err := kit.MaskSensitiveFieldsInMap(configMap)
			if err == nil {
				// 直接使用 map[string]interface{}，structpb.NewStruct 会自动转换
				dto["configJson"] = maskedMap
			}
		}
	}

	return dto
}

// PageModelProvider 获取模型供应器列表（分页）
func (s *ModelService) PageModelProvider(ctx context.Context, req *pb.PageModelProviderRequest) (*pb.Response, error) {
	// 解析过滤条件
	params := &biz.ListModelProviderParams{}
	if req.ModelType != nil && req.ModelType.GetValue() != "" {
		modelType := req.ModelType.GetValue()
		params.ModelType = &modelType
	}
	if req.Name != nil && req.Name.GetValue() != "" {
		name := req.Name.GetValue()
		params.Name = &name
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1 // 默认第1页
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = 10 // 默认10
	}
	page.SetPageNo(int(pageNo)) // 前端传的page从1开始，直接使用
	page.SetPageSize(int(pageSize))

	// 查询列表和总数
	list, total, err := s.providerUC.GetListPage(ctx, params, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.modelProviderToVO(item)
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
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

// AddModelProvider 新增模型供应器
func (s *ModelService) AddModelProvider(ctx context.Context, req *pb.AddModelProviderRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	// 生成ID（UUID去掉横线，32位）
	id := strings.ReplaceAll(uuid.New().String(), "-", "")

	provider := &biz.ModelProvider{
		ID:           id,
		ModelType:    req.GetModelType(),
		ProviderCode: req.GetProviderCode(),
		Name:         req.GetName(),
		Fields:       req.GetFields(),
		Sort:         req.GetSort(),
		Creator:      userID,
		Updater:      userID,
		CreateDate:   time.Now(),
		UpdateDate:   time.Now(),
	}

	created, err := s.providerUC.AddModelProvider(ctx, provider)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 返回DTO
	dto := s.modelProviderToVO(created)
	dataStruct, err := structpb.NewStruct(dto)
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

// EditModelProvider 修改模型供应器
func (s *ModelService) EditModelProvider(ctx context.Context, req *pb.EditModelProviderRequest) (*pb.Response, error) {
	if req == nil {
		return &pb.Response{
			Code: 400,
			Msg:  "参数不能为空",
		}, nil
	}

	// 从context获取用户ID
	userID, _ := middleware.GetUserIdFromContext(ctx)

	provider := &biz.ModelProvider{
		ID:           req.GetId(),
		ModelType:    req.GetModelType(),
		ProviderCode: req.GetProviderCode(),
		Name:         req.GetName(),
		Fields:       req.GetFields(),
		Sort:         req.GetSort(),
		Updater:      userID,
		UpdateDate:   time.Now(),
	}

	updated, err := s.providerUC.EditModelProvider(ctx, provider)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 返回DTO
	dto := s.modelProviderToVO(updated)
	dataStruct, err := structpb.NewStruct(dto)
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

// DeleteModelProvider 删除模型供应器
func (s *ModelService) DeleteModelProvider(ctx context.Context, req *pb.DeleteModelProviderRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "ID列表不能为空",
		}, nil
	}

	err := s.providerUC.DeleteModelProvider(ctx, ids)
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

// GetPluginNames 获取插件名称列表
func (s *ModelService) GetPluginNames(ctx context.Context, req *pb.GetPluginNamesRequest) (*pb.Response, error) {
	list, err := s.providerUC.GetPluginList(ctx)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := s.modelProviderToVO(item)
		voList = append(voList, vo)
	}

	// TODO: 合并知识库数据（如果用户已登录）
	// 参考Java实现，需要查询当前用户的知识库并转换为ModelProvider格式
	// 目前先返回Plugin类型的供应器列表

	data := map[string]interface{}{
		"list": voList,
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

// modelProviderToVO 将biz实体转换为VO
func (s *ModelService) modelProviderToVO(provider *biz.ModelProvider) map[string]interface{} {
	createDate := ""
	if !provider.CreateDate.IsZero() {
		createDate = provider.CreateDate.Format(time.RFC3339)
	}
	updateDate := ""
	if !provider.UpdateDate.IsZero() {
		updateDate = provider.UpdateDate.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"id":           provider.ID,
		"modelType":    provider.ModelType,
		"providerCode": provider.ProviderCode,
		"name":         provider.Name,
		"fields":       provider.Fields, // JSON字符串
		"sort":         provider.Sort,
		"creator":      fmt.Sprintf("%d", provider.Creator),
		"createDate":   createDate,
		"updater":      fmt.Sprintf("%d", provider.Updater),
		"updateDate":   updateDate,
	}
}

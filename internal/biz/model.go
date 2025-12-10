package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// ModelConfig 模型配置实体
type ModelConfig struct {
	ID         string
	ModelType  string
	ModelCode  string
	ModelName  string
	IsDefault  bool
	IsEnabled  bool
	ConfigJSON string // JSON字符串
	DocLink    string
	Remark     string
	Sort       int32
	Creator    int64
	CreateDate time.Time
	Updater    int64
	UpdateDate time.Time
}

// ModelProvider 模型供应器实体
type ModelProvider struct {
	ID           string
	ModelType    string
	ProviderCode string
	Name         string
	Fields       string // JSON字符串
	Sort         int32
	Creator      int64
	CreateDate   time.Time
	Updater      int64
	UpdateDate   time.Time
}

// Voice 音色实体
type Voice struct {
	ID        string
	Name      string
	VoiceDemo string
}

// ListModelConfigParams 查询模型配置过滤条件
type ListModelConfigParams struct {
	ModelType string
	ModelName *string // 可选，模糊查询
}

// ModelConfigRepo 模型配置数据访问接口
type ModelConfigRepo interface {
	GetModelCodeList(ctx context.Context, modelType string, modelName *string) ([]*ModelConfig, error)
	GetLlmModelCodeList(ctx context.Context, modelName *string) ([]*ModelConfig, error)
	GetModelConfigList(ctx context.Context, params *ListModelConfigParams, page *kit.PageRequest) ([]*ModelConfig, error)
	TotalModelConfigs(ctx context.Context, params *ListModelConfigParams) (int, error)
	GetModelConfigByID(ctx context.Context, id string) (*ModelConfig, error)
	GetModelConfigByIDRaw(ctx context.Context, id string) (*ModelConfig, error) // 不经过敏感数据处理
	CreateModelConfig(ctx context.Context, config *ModelConfig) (*ModelConfig, error)
	UpdateModelConfig(ctx context.Context, config *ModelConfig) error
	DeleteModelConfig(ctx context.Context, id string) error
	SetDefaultModel(ctx context.Context, modelType string, isDefault bool) error
	CheckAgentReference(ctx context.Context, modelId string) ([]string, error)    // 返回智能体名称列表
	CheckIntentConfigReference(ctx context.Context, modelId string) (bool, error) // 检查意图识别配置引用
	GetRAGModelList(ctx context.Context) ([]*ModelConfig, error)                  // 获取RAG模型列表
}

// ListModelProviderParams 查询模型供应器过滤条件
type ListModelProviderParams struct {
	ModelType *wrappers.StringValue // 可选，模型类型过滤
	Name      *wrappers.StringValue // 可选，名称或providerCode模糊查询
}

// ModelProviderRepo 模型供应器数据访问接口
type ModelProviderRepo interface {
	GetListByModelType(ctx context.Context, modelType string) ([]*ModelProvider, error)
	GetList(ctx context.Context, modelType string, provideCode string) ([]*ModelProvider, error)
	// 新增方法
	GetListPage(ctx context.Context, params *ListModelProviderParams, page *kit.PageRequest) ([]*ModelProvider, error)
	TotalModelProviders(ctx context.Context, params *ListModelProviderParams) (int, error)
	GetModelProviderByID(ctx context.Context, id string) (*ModelProvider, error)
	CreateModelProvider(ctx context.Context, provider *ModelProvider) (*ModelProvider, error)
	UpdateModelProvider(ctx context.Context, provider *ModelProvider) error
	DeleteModelProvider(ctx context.Context, ids []string) error
	GetPluginList(ctx context.Context) ([]*ModelProvider, error) // 获取Plugin类型的供应器
}

// ModelUsecase 模型配置业务逻辑
type ModelUsecase struct {
	repo         ModelConfigRepo
	providerRepo ModelProviderRepo
	agentRepo    AgentRepo // 用于检查引用
	handleError  *cerrors.HandleError
	log          *log.Helper
}

// NewModelUsecase 创建模型配置用例
func NewModelUsecase(
	repo ModelConfigRepo,
	providerRepo ModelProviderRepo,
	agentRepo AgentRepo,
	logger log.Logger,
) *ModelUsecase {
	return &ModelUsecase{
		repo:         repo,
		providerRepo: providerRepo,
		agentRepo:    agentRepo,
		handleError:  cerrors.NewHandleError(logger),
		log:          kit.LogHelper(logger),
	}
}

// GenerateModelID 生成模型配置ID（UUID去掉横线）
func (uc *ModelUsecase) GenerateModelID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GetModelCodeList 获取模型名称列表
func (uc *ModelUsecase) GetModelCodeList(ctx context.Context, modelType string, modelName *string) ([]*ModelConfig, error) {
	if modelType == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType不能为空"))
	}
	return uc.repo.GetModelCodeList(ctx, modelType, modelName)
}

// GetLlmModelCodeList 获取LLM模型信息列表
func (uc *ModelUsecase) GetLlmModelCodeList(ctx context.Context, modelName *string) ([]*ModelConfig, error) {
	return uc.repo.GetLlmModelCodeList(ctx, modelName)
}

// GetModelProviderList 获取模型供应器列表
func (uc *ModelUsecase) GetModelProviderList(ctx context.Context, modelType string) ([]*ModelProvider, error) {
	if modelType == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType不能为空"))
	}
	return uc.providerRepo.GetListByModelType(ctx, modelType)
}

// GetModelConfigList 获取模型配置列表（分页）
func (uc *ModelUsecase) GetModelConfigList(ctx context.Context, params *ListModelConfigParams, page *kit.PageRequest) ([]*ModelConfig, int, error) {
	if params == nil || params.ModelType == "" {
		return nil, 0, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType不能为空"))
	}
	list, err := uc.repo.GetModelConfigList(ctx, params, page)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.repo.TotalModelConfigs(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetModelConfigByID 获取模型配置详情
func (uc *ModelUsecase) GetModelConfigByID(ctx context.Context, id string) (*ModelConfig, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	config, err := uc.repo.GetModelConfigByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if config == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}
	return config, nil
}

// AddModelConfig 新增模型配置
func (uc *ModelUsecase) AddModelConfig(ctx context.Context, modelType string, provideCode string, config *ModelConfig) (*ModelConfig, error) {
	// 验证参数
	if modelType == "" || provideCode == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType和provideCode不能为空"))
	}
	if config == nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("配置数据不能为空"))
	}

	// 验证供应器是否存在
	providers, err := uc.providerRepo.GetList(ctx, modelType, provideCode)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if len(providers) == 0 {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型供应器不存在"))
	}

	// 如果id为空，自动生成UUID
	if config.ID == "" {
		config.ID = uc.GenerateModelID()
	}

	// 设置默认值
	config.ModelType = modelType
	config.IsDefault = false // 新增时默认非默认
	if config.CreateDate.IsZero() {
		config.CreateDate = time.Now()
	}
	if config.UpdateDate.IsZero() {
		config.UpdateDate = time.Now()
	}

	return uc.repo.CreateModelConfig(ctx, config)
}

// EditModelConfig 编辑模型配置
func (uc *ModelUsecase) EditModelConfig(ctx context.Context, modelType string, provideCode string, id string, config *ModelConfig) (*ModelConfig, error) {
	// 验证参数
	if modelType == "" || provideCode == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType和provideCode不能为空"))
	}
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if config == nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("配置数据不能为空"))
	}

	// 验证供应器是否存在
	providers, err := uc.providerRepo.GetList(ctx, modelType, provideCode)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if len(providers) == 0 {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型供应器不存在"))
	}

	// 从数据库获取原始配置（不经过敏感数据处理）
	originalConfig, err := uc.repo.GetModelConfigByIDRaw(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if originalConfig == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}

	// 验证LLM配置（如果configJson包含llm字段）
	if err := uc.validateLlmConfiguration(ctx, config.ConfigJSON); err != nil {
		return nil, err
	}

	// 处理敏感数据：合并原始配置和更新配置
	mergedConfigJSON, err := uc.mergeConfigJSON(originalConfig.ConfigJSON, config.ConfigJSON)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 准备更新实体
	config.ID = id
	config.ModelType = modelType
	config.ConfigJSON = mergedConfigJSON
	config.UpdateDate = time.Now()

	// 更新数据库
	err = uc.repo.UpdateModelConfig(ctx, config)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 返回更新后的配置（经过敏感数据处理）
	return uc.repo.GetModelConfigByID(ctx, id)
}

// DeleteModelConfig 删除模型配置
func (uc *ModelUsecase) DeleteModelConfig(ctx context.Context, id string) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	// 检查是否为默认模型
	config, err := uc.repo.GetModelConfigByIDRaw(ctx, id)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if config == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}
	if config.IsDefault {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("默认模型不能删除"))
	}

	// 检查智能体引用
	agentNames, err := uc.repo.CheckAgentReference(ctx, id)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if len(agentNames) > 0 {
		agentNamesStr := strings.Join(agentNames, "、")
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("模型被智能体引用，无法删除: %s", agentNamesStr))
	}

	// 检查意图识别配置引用（如果删除的是LLM模型）
	if strings.ToUpper(config.ModelType) == "LLM" {
		hasReference, err := uc.repo.CheckIntentConfigReference(ctx, id)
		if err != nil {
			return uc.handleError.ErrInternal(ctx, err)
		}
		if hasReference {
			return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("LLM模型被意图识别配置引用，无法删除"))
		}
	}

	return uc.repo.DeleteModelConfig(ctx, id)
}

// EnableModelConfig 启用/关闭模型
func (uc *ModelUsecase) EnableModelConfig(ctx context.Context, id string, status bool) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	config, err := uc.repo.GetModelConfigByIDRaw(ctx, id)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if config == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}

	config.IsEnabled = status
	config.ConfigJSON = "" // 不更新ConfigJson字段
	config.UpdateDate = time.Now()

	return uc.repo.UpdateModelConfig(ctx, config)
}

// SetDefaultModel 设置默认模型
func (uc *ModelUsecase) SetDefaultModel(ctx context.Context, id string) (*ModelConfig, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	config, err := uc.repo.GetModelConfigByIDRaw(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if config == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}

	// 将同类型其他模型设置为非默认
	err = uc.repo.SetDefaultModel(ctx, config.ModelType, false)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 设置当前模型为默认
	config.IsDefault = true
	config.IsEnabled = true
	config.ConfigJSON = "" // 不更新ConfigJson字段
	config.UpdateDate = time.Now()

	err = uc.repo.UpdateModelConfig(ctx, config)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 返回更新后的配置
	return uc.repo.GetModelConfigByID(ctx, id)
}

// validateLlmConfiguration 验证LLM配置
func (uc *ModelUsecase) validateLlmConfiguration(ctx context.Context, configJSON string) error {
	if configJSON == "" {
		return nil
	}

	// 解析JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		return nil // JSON解析失败，跳过验证
	}

	// 检查是否包含llm字段
	llmValue, ok := configMap["llm"]
	if !ok {
		return nil
	}

	llmID, ok := llmValue.(string)
	if !ok || llmID == "" {
		return nil
	}

	// 查询LLM模型配置
	llmConfig, err := uc.repo.GetModelConfigByIDRaw(ctx, llmID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if llmConfig == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("LLM模型不存在"))
	}

	// 验证模型类型
	if strings.ToUpper(llmConfig.ModelType) != "LLM" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("LLM模型不存在"))
	}

	// 验证LLM类型（从configJson中提取type字段）
	if llmConfig.ConfigJSON != "" {
		var llmConfigMap map[string]interface{}
		if err := json.Unmarshal([]byte(llmConfig.ConfigJSON), &llmConfigMap); err == nil {
			if typeValue, ok := llmConfigMap["type"].(string); ok {
				if typeValue != "openai" && typeValue != "ollama" {
					return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("无效的LLM类型"))
				}
			}
		}
	}

	return nil
}

// mergeConfigJSON 合并配置JSON，处理敏感数据
func (uc *ModelUsecase) mergeConfigJSON(originalJSON string, updatedJSON string) (string, error) {
	if updatedJSON == "" {
		return originalJSON, nil
	}
	if originalJSON == "" {
		return updatedJSON, nil
	}

	// 解析原始JSON
	var originalMap map[string]interface{}
	if err := json.Unmarshal([]byte(originalJSON), &originalMap); err != nil {
		return updatedJSON, nil // 解析失败，使用更新值
	}

	// 解析更新JSON
	var updatedMap map[string]interface{}
	if err := json.Unmarshal([]byte(updatedJSON), &updatedMap); err != nil {
		return originalJSON, nil // 解析失败，使用原始值
	}

	// 合并JSON
	resultMap := make(map[string]interface{})
	for k, v := range originalMap {
		resultMap[k] = v
	}

	for key, value := range updatedMap {
		keyLower := strings.ToLower(key)
		if kit.IsSensitiveField(keyLower) {
			// 敏感字段：只更新非掩码值
			if strValue, ok := value.(string); ok && !kit.IsMaskedValue(strValue) {
				resultMap[key] = value
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套JSON
			originalNested, _ := resultMap[key].(map[string]interface{})
			if originalNested == nil {
				originalNested = make(map[string]interface{})
			}
			mergedNested, err := uc.mergeNestedJSON(originalNested, nestedMap)
			if err == nil {
				resultMap[key] = mergedNested
			}
		} else {
			// 非敏感字段直接更新
			resultMap[key] = value
		}
	}

	// 转换回JSON字符串
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		return originalJSON, err
	}

	return string(resultJSON), nil
}

// mergeNestedJSON 递归合并嵌套JSON
func (uc *ModelUsecase) mergeNestedJSON(original map[string]interface{}, updated map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range original {
		result[k] = v
	}

	for key, value := range updated {
		keyLower := strings.ToLower(key)
		if kit.IsSensitiveField(keyLower) {
			// 敏感字段：只更新非掩码值
			if strValue, ok := value.(string); ok && !kit.IsMaskedValue(strValue) {
				result[key] = value
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套JSON
			originalNested, _ := result[key].(map[string]interface{})
			if originalNested == nil {
				originalNested = make(map[string]interface{})
			}
			mergedNested, err := uc.mergeNestedJSON(originalNested, nestedMap)
			if err == nil {
				result[key] = mergedNested
			}
		} else {
			result[key] = value
		}
	}

	return result, nil
}

// ModelProviderUsecase 模型供应器业务逻辑
type ModelProviderUsecase struct {
	repo        ModelProviderRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewModelProviderUsecase 创建模型供应器用例
func NewModelProviderUsecase(
	repo ModelProviderRepo,
	logger log.Logger,
) *ModelProviderUsecase {
	return &ModelProviderUsecase{
		repo:        repo,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// GetListPage 分页查询模型供应器
func (uc *ModelProviderUsecase) GetListPage(ctx context.Context, params *ListModelProviderParams, page *kit.PageRequest) ([]*ModelProvider, int, error) {
	if err := kit.Validate(params); err != nil {
		return nil, 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	list, err := uc.repo.GetListPage(ctx, params, page)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.repo.TotalModelProviders(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AddModelProvider 新增模型供应器
func (uc *ModelProviderUsecase) AddModelProvider(ctx context.Context, provider *ModelProvider) (*ModelProvider, error) {
	if err := kit.Validate(provider); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 验证必填字段
	if provider.ModelType == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType不能为空"))
	}
	if provider.ProviderCode == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("providerCode不能为空"))
	}
	if provider.Name == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("name不能为空"))
	}
	if provider.Fields == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("fields不能为空"))
	}

	// 设置默认值
	if provider.CreateDate.IsZero() {
		provider.CreateDate = time.Now()
	}
	if provider.UpdateDate.IsZero() {
		provider.UpdateDate = time.Now()
	}

	return uc.repo.CreateModelProvider(ctx, provider)
}

// EditModelProvider 修改模型供应器
func (uc *ModelProviderUsecase) EditModelProvider(ctx context.Context, provider *ModelProvider) (*ModelProvider, error) {
	if err := kit.Validate(provider); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 验证必填字段
	if provider.ID == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if provider.ModelType == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelType不能为空"))
	}
	if provider.ProviderCode == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("providerCode不能为空"))
	}
	if provider.Name == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("name不能为空"))
	}
	if provider.Fields == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("fields不能为空"))
	}

	// 检查供应器是否存在
	existing, err := uc.repo.GetModelProviderByID(ctx, provider.ID)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if existing == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型供应器不存在"))
	}

	// 设置更新时间
	provider.UpdateDate = time.Now()

	err = uc.repo.UpdateModelProvider(ctx, provider)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 返回更新后的供应器
	return uc.repo.GetModelProviderByID(ctx, provider.ID)
}

// DeleteModelProvider 批量删除模型供应器
func (uc *ModelProviderUsecase) DeleteModelProvider(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("ID列表不能为空"))
	}

	return uc.repo.DeleteModelProvider(ctx, ids)
}

// GetPluginList 获取插件列表（需要合并知识库数据）
func (uc *ModelProviderUsecase) GetPluginList(ctx context.Context) ([]*ModelProvider, error) {
	// 获取Plugin类型的供应器
	list, err := uc.repo.GetPluginList(ctx)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// TODO: 合并知识库数据（如果用户已登录）
	// 参考Java实现，需要查询当前用户的知识库并转换为ModelProvider格式
	// 目前先返回Plugin类型的供应器列表

	return list, nil
}

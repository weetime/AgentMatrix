package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// SysParam 系统参数
type SysParam struct {
	ID         int64
	ParamCode  string
	ParamValue string
	ValueType  string
	ParamType  int8
	Remark     string
	Creator    int64
	Updater    int64
}

// ListSysParamsParams 查询参数过滤条件
type ListSysParamsParams struct {
	ParamCode *wrappers.StringValue // 参数编码或备注（模糊查询）
}

// ConfigRepo 配置数据访问接口
type ConfigRepo interface {
	ListAllParams(ctx context.Context) ([]*SysParam, error)
	// 分页查询参数（使用项目规范）
	ListSysParams(ctx context.Context, params *ListSysParamsParams, page *kit.PageRequest) ([]*SysParam, error)
	TotalSysParams(ctx context.Context, params *ListSysParamsParams) (int, error)
	// 兼容旧接口
	PageSysParams(ctx context.Context, page, limit int32, paramCode string) ([]*SysParam, int, error)
	GetSysParamsByID(ctx context.Context, id int64) (*SysParam, error)
	CreateSysParams(ctx context.Context, param *SysParam) (*SysParam, error)
	UpdateSysParams(ctx context.Context, param *SysParam) error
	DeleteSysParams(ctx context.Context, ids []int64) error
	GetSysParamsByCode(ctx context.Context, paramCode string) (*SysParam, error)
}

// ConfigUsecase 配置业务逻辑
type ConfigUsecase struct {
	repo        ConfigRepo
	agentRepo   AgentRepo
	modelRepo   ModelConfigRepo
	redisClient *kit.RedisClient
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewConfigUsecase 创建配置用例
func NewConfigUsecase(
	repo ConfigRepo,
	agentRepo AgentRepo,
	modelRepo ModelConfigRepo,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *ConfigUsecase {
	return &ConfigUsecase{
		repo:        repo,
		agentRepo:   agentRepo,
		modelRepo:   modelRepo,
		redisClient: redisClient,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// GetConfig 获取服务器配置
// isCache: 是否从缓存读取
func (uc *ConfigUsecase) GetConfig(ctx context.Context, isCache bool) (map[string]interface{}, error) {
	// 如果启用缓存，先从 Redis 读取
	if isCache {
		var cachedConfig map[string]interface{}
		err := uc.redisClient.GetObject(ctx, kit.RedisKeyServerConfig, &cachedConfig)
		if err == nil && cachedConfig != nil && len(cachedConfig) > 0 {
			uc.log.Debug("Config loaded from cache")
			// 清理 Java 类型信息
			cleaned := uc.cleanJavaClassInfo(cachedConfig)
			if cleanedConfig, ok := cleaned.(map[string]interface{}); ok {
				return cleanedConfig, nil
			}
			// 如果类型不匹配，返回原配置
			return cachedConfig, nil
		}
		// 缓存不存在或为空时，继续构建配置（不报错）
		if err != nil {
			uc.log.Debug("Cache not found or empty, will build config from database", "error", err)
		} else {
			uc.log.Debug("Cache not found or empty, will build config from database")
		}
	}

	// 构建配置
	config, err := uc.buildConfig(ctx)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 查询默认智能体模板（sort=1）并构建模块配置
	templates, err := uc.agentRepo.GetAgentTemplateList(ctx)
	if err == nil && len(templates) > 0 {
		// 找到 sort=1 的模板，如果没有则使用第一个
		var defaultTemplate *AgentTemplate
		for _, template := range templates {
			if template.Sort == 1 {
				defaultTemplate = template
				break
			}
		}
		if defaultTemplate == nil && len(templates) > 0 {
			defaultTemplate = templates[0]
		}

		if defaultTemplate != nil {
			// 构建模块配置
			if err := uc.buildModuleConfig(ctx, defaultTemplate, config); err != nil {
				uc.log.Warn("Failed to build module config", "error", err)
			}
		}
	} else if err != nil {
		uc.log.Warn("Failed to get agent templates", "error", err)
	}

	// 清理 Java 类型信息（防止之前缓存的数据污染）
	cleaned := uc.cleanJavaClassInfo(config)
	cleanedConfig, ok := cleaned.(map[string]interface{})
	if !ok {
		// 如果类型不匹配，使用原配置
		cleanedConfig = config
	}

	// 将配置存入 Redis（不设置过期时间，永久缓存）
	if err := uc.redisClient.SetObject(ctx, kit.RedisKeyServerConfig, cleanedConfig, 0); err != nil {
		uc.log.Warn("Failed to save config to cache", "error", err)
	}

	return cleanedConfig, nil
}

// buildConfig 构建配置信息
// 从 sys_params 表读取所有参数，按照 param_code 的点号分隔构建嵌套 Map
func (uc *ConfigUsecase) buildConfig(ctx context.Context) (map[string]interface{}, error) {
	// 查询所有系统参数（param_type = 1）
	params, err := uc.repo.ListAllParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list params: %w", err)
	}

	config := make(map[string]interface{})

	for _, param := range params {
		// 按照 param_code 的点号分隔构建嵌套 Map
		keys := strings.Split(param.ParamCode, ".")
		current := config

		// 遍历除最后一个 key 之外的所有 key，创建嵌套 Map
		for i := 0; i < len(keys)-1; i++ {
			key := keys[i]
			if _, exists := current[key]; !exists {
				current[key] = make(map[string]interface{})
			}
			// 类型断言，确保是 map[string]interface{}
			if nextMap, ok := current[key].(map[string]interface{}); ok {
				current = nextMap
			} else {
				// 如果类型不匹配，创建新的 map
				current[key] = make(map[string]interface{})
				current = current[key].(map[string]interface{})
			}
		}

		// 处理最后一个 key，根据 value_type 转换值
		lastKey := keys[len(keys)-1]
		value := uc.convertValue(param.ParamValue, param.ValueType)
		current[lastKey] = value
	}

	return config, nil
}

// convertValue 根据 value_type 转换值
func (uc *ConfigUsecase) convertValue(value, valueType string) interface{} {
	switch strings.ToLower(valueType) {
	case "number":
		// 尝试解析为数字
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			// 如果是整数形式，转换为 int
			if num == float64(int64(num)) {
				return int64(num)
			}
			return num
		}
		// 解析失败，返回原字符串
		return value

	case "boolean":
		// 解析为布尔值
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
		return value

	case "array":
		// 将分号分隔的字符串转换为字符串数组
		// 注意：返回 []interface{} 而不是 []string，因为 structpb.NewStruct 不支持 []string
		parts := strings.Split(value, ";")
		result := make([]interface{}, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result

	case "json":
		// 解析为 JSON 对象
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(value), &jsonObj); err == nil {
			return jsonObj
		}
		// 解析失败，返回原字符串
		return value

	default:
		// string 类型或其他未知类型，直接返回字符串
		return value
	}
}

// cleanJavaClassInfo 清理 Java 类型信息（@class 字段）
// 递归清理 map 和 slice 中的 @class 字段
func (uc *ConfigUsecase) cleanJavaClassInfo(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// 跳过 @class 字段
			if key == "@class" {
				continue
			}
			// 递归清理嵌套结构
			result[key] = uc.cleanJavaClassInfo(value)
		}
		return result

	case []interface{}:
		// 处理特殊格式：["java.util.ArrayList", [实际数据]]
		if len(v) == 2 {
			if typeStr, ok := v[0].(string); ok && strings.HasPrefix(typeStr, "java.util.") {
				if actualArray, ok := v[1].([]interface{}); ok {
					// 返回实际的数组，并递归清理
					return uc.cleanJavaClassInfo(actualArray)
				}
			}
		}
		// 普通数组处理
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			// 跳过 Java 类型字符串（如 "java.util.ArrayList"）
			if str, ok := item.(string); ok && strings.HasPrefix(str, "java.util.") {
				continue
			}
			// 递归清理嵌套结构
			result = append(result, uc.cleanJavaClassInfo(item))
		}
		return result

	case []string:
		// 处理字符串数组，跳过 Java 类型字符串
		// 转换为 []interface{} 以支持 structpb.NewStruct
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			if !strings.HasPrefix(item, "java.util.") {
				result = append(result, item)
			}
		}
		return result

	case map[string]string:
		// 处理 map[string]string，转换为 map[string]interface{} 以支持 structpb.NewStruct
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = value
		}
		return result

	default:
		// 其他类型直接返回
		return data
	}
}

// ListSysParams 列表查询参数（使用项目规范）
func (uc *ConfigUsecase) ListSysParams(ctx context.Context, params *ListSysParamsParams, page *kit.PageRequest) ([]*SysParam, error) {
	if err := kit.Validate(params); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.ListSysParams(ctx, params, page)
}

// TotalSysParams 获取参数总数
func (uc *ConfigUsecase) TotalSysParams(ctx context.Context, params *ListSysParamsParams) (int, error) {
	if err := kit.Validate(params); err != nil {
		return 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	return uc.repo.TotalSysParams(ctx, params)
}

// PageSysParams 分页查询参数（兼容旧接口）
func (uc *ConfigUsecase) PageSysParams(ctx context.Context, page, limit int32, paramCode string) ([]*SysParam, int, error) {
	return uc.repo.PageSysParams(ctx, page, limit, paramCode)
}

// GetSysParamsByID 根据ID获取参数
func (uc *ConfigUsecase) GetSysParamsByID(ctx context.Context, id int64) (*SysParam, error) {
	return uc.repo.GetSysParamsByID(ctx, id)
}

// CreateSysParams 创建参数
func (uc *ConfigUsecase) CreateSysParams(ctx context.Context, param *SysParam) (*SysParam, error) {
	// 检查参数编码是否已存在
	existing, err := uc.repo.GetSysParamsByCode(ctx, param.ParamCode)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("参数编码 %s 已存在", param.ParamCode)
	}

	// 设置默认值
	if param.ValueType == "" {
		param.ValueType = "string"
	}
	if param.ParamType == 0 {
		param.ParamType = 1 // 默认为非系统参数
	}

	return uc.repo.CreateSysParams(ctx, param)
}

// UpdateSysParams 更新参数
func (uc *ConfigUsecase) UpdateSysParams(ctx context.Context, param *SysParam) error {
	// 检查参数是否存在
	existing, err := uc.repo.GetSysParamsByID(ctx, param.ID)
	if err != nil {
		return fmt.Errorf("参数不存在: %w", err)
	}

	// 如果修改了参数编码，检查新编码是否已存在
	if param.ParamCode != existing.ParamCode {
		codeExists, err := uc.repo.GetSysParamsByCode(ctx, param.ParamCode)
		if err == nil && codeExists != nil && codeExists.ID != param.ID {
			return fmt.Errorf("参数编码 %s 已存在", param.ParamCode)
		}
	}

	// 验证参数值
	if err := uc.validateParamValue(ctx, param.ParamCode, param.ParamValue); err != nil {
		return err
	}

	// 更新参数
	if err := uc.repo.UpdateSysParams(ctx, param); err != nil {
		return err
	}

	// 清除配置缓存
	if uc.redisClient != nil {
		uc.redisClient.Delete(ctx, kit.RedisKeyServerConfig)
	}

	return nil
}

// DeleteSysParams 批量删除参数
func (uc *ConfigUsecase) DeleteSysParams(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("参数ID列表不能为空")
	}

	if err := uc.repo.DeleteSysParams(ctx, ids); err != nil {
		return err
	}

	// 清除配置缓存
	if uc.redisClient != nil {
		uc.redisClient.Delete(ctx, kit.RedisKeyServerConfig)
	}

	return nil
}

// validateParamValue 验证参数值
func (uc *ConfigUsecase) validateParamValue(ctx context.Context, paramCode, paramValue string) error {
	// 参数编码常量
	const (
		SERVER_WEBSOCKET     = "server.websocket"
		SERVER_SECRET        = "server.secret"
		SERVER_OTA           = "server.ota"
		SERVER_MCP_ENDPOINT  = "server.mcp_endpoint"
		SERVER_VOICE_PRINT   = "server.voice_print"
		SERVER_MQTT_SIGN_KEY = "server.mqtt_signature_key"
	)

	switch paramCode {
	case SERVER_WEBSOCKET:
		return uc.validateWebSocketUrl(paramValue)
	case SERVER_OTA:
		return uc.validateOtaUrl(paramValue)
	case SERVER_MCP_ENDPOINT:
		return uc.validateMcpUrl(paramValue)
	case SERVER_VOICE_PRINT:
		return uc.validateVoicePrintUrl(paramValue)
	case SERVER_MQTT_SIGN_KEY:
		return uc.validateMqttKey(paramValue)
	default:
		return nil
	}
}

// validateWebSocketUrl 验证WebSocket地址
func (uc *ConfigUsecase) validateWebSocketUrl(url string) error {
	if url == "" || url == "null" {
		return nil
	}

	// 检查是否包含localhost或127.0.0.1
	if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
		return fmt.Errorf("WebSocket地址不能包含localhost或127.0.0.1")
	}

	// 验证格式：ws://或wss://开头
	if !strings.HasPrefix(url, "ws://") && !strings.HasPrefix(url, "wss://") {
		return fmt.Errorf("WebSocket地址格式错误，必须以ws://或wss://开头")
	}

	// TODO: 连接测试（需要实现WebSocket客户端测试）
	// 这里暂时只做格式验证

	return nil
}

// validateOtaUrl 验证OTA地址
func (uc *ConfigUsecase) validateOtaUrl(url string) error {
	if url == "" || url == "null" {
		return nil
	}

	// 检查是否包含localhost或127.0.0.1
	if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
		return fmt.Errorf("OTA地址不能包含localhost或127.0.0.1")
	}

	// 必须以http开头
	if !strings.HasPrefix(strings.ToLower(url), "http") {
		return fmt.Errorf("OTA地址必须以http开头")
	}

	// 必须以/ota/结尾
	if !strings.HasSuffix(url, "/ota/") {
		return fmt.Errorf("OTA地址必须以/ota/结尾")
	}

	// TODO: 连接测试（需要实现HTTP请求测试）

	return nil
}

// validateMcpUrl 验证MCP地址
func (uc *ConfigUsecase) validateMcpUrl(url string) error {
	if url == "" || url == "null" {
		return nil
	}

	// 检查是否包含localhost或127.0.0.1
	if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
		return fmt.Errorf("MCP地址不能包含localhost或127.0.0.1")
	}

	// 必须包含"key"
	if !strings.Contains(url, "key") {
		return fmt.Errorf("MCP地址必须包含key")
	}

	// TODO: 连接测试（需要实现HTTP请求测试）

	return nil
}

// validateVoicePrintUrl 验证声纹地址
func (uc *ConfigUsecase) validateVoicePrintUrl(url string) error {
	if url == "" || url == "null" {
		return nil
	}

	// 检查是否包含localhost或127.0.0.1
	if strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1") {
		return fmt.Errorf("声纹地址不能包含localhost或127.0.0.1")
	}

	// 必须以http开头
	if !strings.HasPrefix(strings.ToLower(url), "http") {
		return fmt.Errorf("声纹地址必须以http开头")
	}

	// 必须包含"key"
	if !strings.Contains(url, "key") {
		return fmt.Errorf("声纹地址必须包含key")
	}

	// TODO: 健康检查（需要实现HTTP请求测试）

	return nil
}

// validateMqttKey 验证MQTT密钥
func (uc *ConfigUsecase) validateMqttKey(key string) error {
	if key == "" || key == "null" {
		return nil
	}

	// 长度至少8位
	if len(key) < 8 {
		return fmt.Errorf("MQTT密钥长度至少8位")
	}

	// 包含大小写字母
	hasUpper := false
	hasLower := false
	for _, r := range key {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if r >= 'a' && r <= 'z' {
			hasLower = true
		}
	}
	if !hasUpper || !hasLower {
		return fmt.Errorf("MQTT密钥必须包含大小写字母")
	}

	// 弱密码检查
	weakPasswords := []string{"123456", "password", "admin", "qwerty", "abc123"}
	keyLower := strings.ToLower(key)
	for _, weak := range weakPasswords {
		if strings.Contains(keyLower, weak) {
			return fmt.Errorf("MQTT密钥不能包含弱密码")
		}
	}

	return nil
}

// ValidateParamValue 验证参数值（公开方法供service调用）
func (uc *ConfigUsecase) ValidateParamValue(ctx context.Context, paramCode, paramValue string) error {
	return uc.validateParamValue(ctx, paramCode, paramValue)
}

// ClearConfigCache 清除配置缓存
func (uc *ConfigUsecase) ClearConfigCache(ctx context.Context) {
	if uc.redisClient != nil {
		uc.redisClient.Delete(ctx, kit.RedisKeyServerConfig)
	}
}

// GetValue 获取参数值
func (uc *ConfigUsecase) GetValue(ctx context.Context, paramCode string, isCache bool) (string, error) {
	param, err := uc.repo.GetSysParamsByCode(ctx, paramCode)
	if err != nil {
		return "", err
	}
	if param == nil {
		return "", nil
	}
	return param.ParamValue, nil
}

// GetValueObject 获取参数值并转换为对象
func (uc *ConfigUsecase) GetValueObject(ctx context.Context, paramCode string, isCache bool, dest interface{}) error {
	value, err := uc.GetValue(ctx, paramCode, isCache)
	if err != nil {
		return err
	}
	if value == "" {
		return nil
	}
	// 简单的字符串到bool/int转换
	if boolPtr, ok := dest.(*bool); ok {
		*boolPtr = value == "true"
		return nil
	}
	if intPtr, ok := dest.(*int); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			*intPtr = intVal
		}
		return nil
	}
	return nil
}

// buildModuleConfig 构建模块配置（参考Java的buildModuleConfig方法）
func (uc *ConfigUsecase) buildModuleConfig(ctx context.Context, template *AgentTemplate, config map[string]interface{}) error {
	// 使用 map[string]interface{} 而不是 map[string]string，以支持 structpb.NewStruct
	selectedModule := make(map[string]interface{})

	// 定义模型类型和对应的模型ID（只处理 VAD 和 ASR，与 Java 项目保持一致）
	modelTypes := []string{"VAD", "ASR"}
	modelIds := []string{
		template.VADModelID,
		template.ASRModelID,
	}

	for i, modelId := range modelIds {
		if modelId == "" {
			continue
		}

		modelType := modelTypes[i]

		// 获取模型配置
		model, err := uc.modelRepo.GetModelConfigByID(ctx, modelId)
		if err != nil || model == nil {
			uc.log.Warn("Model config not found", "modelId", modelId, "modelType", modelType, "error", err)
			continue
		}

		// 解析模型配置JSON
		var configJSON map[string]interface{}
		if model.ConfigJSON != "" {
			if err := json.Unmarshal([]byte(model.ConfigJSON), &configJSON); err != nil {
				uc.log.Warn("Failed to parse model config JSON", "modelId", modelId, "error", err)
				continue
			}
		} else {
			configJSON = make(map[string]interface{})
		}

		// 构建类型配置
		typeConfig := make(map[string]interface{})
		typeConfig[model.ID] = configJSON

		// 将类型配置添加到结果中
		if existing, ok := config[modelType].(map[string]interface{}); ok {
			// 如果已存在，合并配置
			for k, v := range typeConfig {
				existing[k] = v
			}
			config[modelType] = existing
		} else {
			config[modelType] = typeConfig
		}

		// 添加到selected_module（值作为字符串存储）
		selectedModule[modelType] = model.ID
	}

	// 添加selected_module到配置
	config["selected_module"] = selectedModule

	// 设置 prompt 和 summaryMemory 为 null（与 Java 项目保持一致）
	config["prompt"] = nil
	config["summaryMemory"] = nil

	return nil
}

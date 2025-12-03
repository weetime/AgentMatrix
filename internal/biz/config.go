package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// SysParam 系统参数
type SysParam struct {
	ID        int64
	ParamCode string
	ParamValue string
	ValueType string
	ParamType int8
	Remark    string
}

// ConfigRepo 配置数据访问接口
type ConfigRepo interface {
	ListAllParams(ctx context.Context) ([]*SysParam, error)
}

// ConfigUsecase 配置业务逻辑
type ConfigUsecase struct {
	repo        ConfigRepo
	redisClient *kit.RedisClient
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewConfigUsecase 创建配置用例
func NewConfigUsecase(
	repo ConfigRepo,
	redisClient *kit.RedisClient,
	logger log.Logger,
) *ConfigUsecase {
	return &ConfigUsecase{
		repo:        repo,
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
		if err == nil && cachedConfig != nil {
			uc.log.Debug("Config loaded from cache")
			return cachedConfig, nil
		}
		if err != nil {
			uc.log.Warn("Failed to load config from cache", "error", err)
		}
	}

	// 构建配置
	config, err := uc.buildConfig(ctx)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 将配置存入 Redis（不设置过期时间，永久缓存）
	if err := uc.redisClient.SetObject(ctx, kit.RedisKeyServerConfig, config, 0); err != nil {
		uc.log.Warn("Failed to save config to cache", "error", err)
	}

	return config, nil
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
		parts := strings.Split(value, ";")
		result := make([]string, 0, len(parts))
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


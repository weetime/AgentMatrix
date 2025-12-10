package kit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客户端封装
type RedisClient struct {
	client *redis.Client
	log    *log.Helper
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(conf *conf.Bootstrap, logger log.Logger) (*RedisClient, func(), error) {
	log := log.NewHelper(log.With(logger, "module", "agent-matrix-service/redis"))

	opts := &redis.Options{
		Addr:         conf.Data.Redis.Addr,
		ReadTimeout:  conf.Data.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: conf.Data.Redis.WriteTimeout.AsDuration(),
		// 默认使用 DB 0，如果需要可以扩展配置
		DB: 0,
	}

	// 如果配置了密码，则设置密码
	if password := conf.Data.Redis.GetPassword(); password != "" {
		opts.Password = password
		log.Debug("Redis password configured")
	}

	rdb := redis.NewClient(opts)

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("Failed to connect to Redis", "addr", conf.Data.Redis.Addr, "error", err)
		return nil, nil, fmt.Errorf("failed to connect to Redis at %s: %w", conf.Data.Redis.Addr, err)
	}

	log.Info("Redis connection established", "addr", conf.Data.Redis.Addr)

	client := &RedisClient{
		client: rdb,
		log:    log,
	}

	cleanup := func() {
		log.Info("Closing Redis connection")
		if err := rdb.Close(); err != nil {
			log.Error("Failed to close Redis connection", "error", err)
		}
	}

	return client, cleanup, nil
}

// Get 获取字符串值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set 设置字符串值
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(data)
	}
	return r.client.Set(ctx, key, val, expiration).Err()
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetObject 获取对象（JSON 反序列化）
func (r *RedisClient) GetObject(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// SetObject 设置对象（JSON 序列化）
func (r *RedisClient) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.Set(ctx, key, string(data), expiration)
}

// GetClient 获取底层的redis.Client（用于需要直接访问redis.Client的场景）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// RedisKeys Redis Key 常量
const (
	RedisKeyServerConfig = "server:config" // 服务器配置缓存 Key
	RedisKeySysParams    = "sys:params"    // 系统参数缓存 Key
)

// GetDictDataByTypeKey 获取字典数据的缓存key
func GetDictDataByTypeKey(dictType string) string {
	return fmt.Sprintf("sys:dict:data:%s", dictType)
}

// GetTimbreNameByIdKey 获取音色名称缓存key
func GetTimbreNameByIdKey(id string) string {
	return fmt.Sprintf("timbre:name:%s", id)
}

// GetVoiceCloneAudioIdKey 获取音色克隆音频ID的缓存key
func GetVoiceCloneAudioIdKey(uuid string) string {
	return fmt.Sprintf("voiceClone:audio:id:%s", uuid)
}

// GetOtaIdKey 获取OTA固件ID的缓存key
func GetOtaIdKey(uuid string) string {
	return fmt.Sprintf("ota:id:%s", uuid)
}

// GetOtaDownloadCountKey 获取OTA固件下载次数的缓存key
func GetOtaDownloadCountKey(uuid string) string {
	return fmt.Sprintf("ota:download:count:%s", uuid)
}

// GetRedisObject 获取Redis对象（辅助函数，用于直接使用redis.Client的场景）
func GetRedisObject(ctx context.Context, client *redis.Client, key string, dest interface{}) error {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// SetRedisObject 设置Redis对象（辅助函数，用于直接使用redis.Client的场景）
func SetRedisObject(ctx context.Context, client *redis.Client, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return client.Set(ctx, key, string(data), expiration).Err()
}

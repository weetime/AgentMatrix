package kit

import (
	"sync"

	"github.com/sony/sonyflake"
)

// IDGenerator ID生成器（基于Sonyflake）
type IDGenerator struct {
	sf *sonyflake.Sonyflake
}

var (
	defaultIDGenerator *IDGenerator
	once               sync.Once
)

// GetIDGenerator 获取默认的ID生成器单例
func GetIDGenerator() *IDGenerator {
	once.Do(func() {
		sf := sonyflake.NewSonyflake(sonyflake.Settings{})
		defaultIDGenerator = &IDGenerator{
			sf: sf,
		}
	})
	return defaultIDGenerator
}

// GenerateID 生成唯一ID（基于Sonyflake）
func (g *IDGenerator) GenerateID() int64 {
	id, _ := g.sf.NextID()
	return int64(id)
}

// GenerateInt64ID 生成int64类型的ID（便捷方法）
func GenerateInt64ID() int64 {
	return GetIDGenerator().GenerateID()
}

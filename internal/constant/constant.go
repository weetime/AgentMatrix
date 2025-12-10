package constant

// 音色克隆相关常量
const (
	// VoiceClonePrefix 复刻音色前缀（对应 Java 的 ErrorCode.VOICE_CLONE_PREFIX = 10158）
	// 注意：如果后续需要国际化支持，可以改为从国际化资源中获取
	VoiceClonePrefix = "复刻音色"
	MaxUploadSize    = 10 * 1024 * 1024 // 10MB 文件最大上传大小
)

package constant

// 音色克隆相关常量
const (
	// VoiceClonePrefix 复刻音色前缀（对应 Java 的 ErrorCode.VOICE_CLONE_PREFIX = 10158）
	// 注意：如果后续需要国际化支持，可以改为从国际化资源中获取
	VoiceClonePrefix = "复刻音色"
	MaxUploadSize    = 10 * 1024 * 1024 // 10MB 文件最大上传大小
)

// 记忆模型相关常量
const (
	// MemoryNoMem 无记忆模型标识（对应 Java 的 Constant.MEMORY_NO_MEM）
	MemoryNoMem = "Memory_nomem"
)

// 声音克隆类型常量
const (
	// VoiceCloneHuoshanDoubleStream 火山引擎双声道语音克隆（对应 Java 的 Constant.VOICE_CLONE_HUOSHAN_DOUBLE_STREAM）
	VoiceCloneHuoshanDoubleStream = "huoshan_double_stream"
)

// ChatHistoryConf 聊天记录配置枚举（对应 Java 的 Constant.ChatHistoryConfEnum）
const (
	ChatHistoryConfIgnore          = 0 // 不记录
	ChatHistoryConfRecordText      = 1 // 记录文本
	ChatHistoryConfRecordTextAudio = 2 // 文本音频都记录
)

// 错误码常量（对应 Java 的 ErrorCode）
const (
	ErrorCodeChatHistoryNoPermission = 10132 // 没有权限查看该智能体的聊天记录
	ErrorCodeDownloadLinkExpired     = 10136 // 下载链接已过期或无效
	ErrorCodeDownloadLinkInvalid     = 10137 // 下载链接无效
)

# Manager-API 功能清单

本文档详细列出了 manager-api 项目的所有功能，特别是 API 端点，用于指导从 Java 迁移到 Golang 的开发工作。

## 项目基本信息

- **技术栈**: Spring Boot 3.4.3, Java 21, MyBatis Plus, Apache Shiro, MySQL, Redis
- **端口**: 8002
- **上下文路径**: `/xiaozhi`
- **API 文档**: http://localhost:8002/xiaozhi/doc.html
- **数据库**: MySQL 8.0+
- **缓存**: Redis 5.0+

---

## 1. 认证与授权模块 (Security)

### 1.1 登录管理 (`/user`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/user/captcha` | 获取图形验证码 | 公开 | 需要传入 uuid 参数 |
| POST | `/user/smsVerification` | 发送短信验证码 | 公开 | 需要先验证图形验证码 |
| POST | `/user/login` | 用户登录 | 公开 | 支持 SM2 加密密码，返回 TokenDTO |
| POST | `/user/register` | 用户注册 | 公开 | 支持手机号注册，需要短信验证码 |
| GET | `/user/info` | 获取当前用户信息 | 普通用户 | 返回 UserDetail |
| PUT | `/user/change-password` | 修改密码 | 普通用户 | 需要旧密码和新密码 |
| PUT | `/user/retrieve-password` | 找回密码 | 公开 | 需要手机号和短信验证码 |
| GET | `/user/pub-config` | 获取公共配置 | 公开 | 返回系统公共配置（版本、备案号、SM2公钥等） |

**关键特性**:
- SM2 国密加密支持
- 图形验证码 + 短信验证码双重验证
- 手机号注册支持
- JWT Token 认证

---

## 2. 智能体管理模块 (Agent)

### 2.1 智能体管理 (`/agent`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/agent/list` | 获取用户智能体列表 | 普通用户 | 返回当前用户的所有智能体 |
| GET | `/agent/all` | 智能体列表（管理员） | 超级管理员 | 分页查询所有智能体 |
| GET | `/agent/{id}` | 获取智能体详情 | 普通用户 | 返回智能体完整信息 |
| POST | `/agent` | 创建智能体 | 普通用户 | 创建新的智能体 |
| PUT | `/agent/saveMemory/{macAddress}` | 根据设备更新智能体记忆 | 公开 | 通过设备 MAC 地址更新智能体记忆 |
| PUT | `/agent/{id}` | 更新智能体 | 普通用户 | 更新智能体信息 |
| DELETE | `/agent/{id}` | 删除智能体 | 普通用户 | 级联删除关联设备、聊天记录、插件 |
| GET | `/agent/template` | 获取智能体模板列表 | 普通用户 | 返回所有可用的智能体模板 |
| GET | `/agent/{id}/sessions` | 获取智能体会话列表 | 普通用户 | 分页查询智能体的所有会话 |
| GET | `/agent/{id}/chat-history/{sessionId}` | 获取智能体聊天记录 | 普通用户 | 获取指定会话的聊天记录 |
| GET | `/agent/{id}/chat-history/user` | 获取智能体最近50条聊天记录 | 普通用户 | 用户视角的聊天记录 |
| GET | `/agent/{id}/chat-history/audio` | 获取音频内容 | 普通用户 | 根据音频ID获取文本内容 |
| POST | `/agent/audio/{audioId}` | 获取音频下载ID | 普通用户 | 生成临时下载链接 |
| GET | `/agent/play/{uuid}` | 播放音频 | 公开 | 通过临时UUID播放音频文件 |

### 2.2 智能体模板管理 (`/agent/template`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/agent/template/page` | 分页查询模板 | 超级管理员 | 支持按名称模糊查询 |
| GET | `/agent/template/{id}` | 获取模板详情 | 超级管理员 | 返回模板完整信息 |
| POST | `/agent/template` | 创建模板 | 超级管理员 | 自动设置排序值 |
| PUT | `/agent/template` | 更新模板 | 超级管理员 | 更新模板信息 |
| DELETE | `/agent/template/{id}` | 删除模板 | 超级管理员 | 删除后自动重新排序 |
| POST | `/agent/template/batch-remove` | 批量删除模板 | 超级管理员 | 批量删除多个模板 |

### 2.3 智能体聊天历史管理 (`/agent/chat-history`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| POST | `/agent/chat-history/report` | 小智服务聊天上报 | 公开 | 接收 xiaozhi-server 上报的聊天记录 |
| POST | `/agent/chat-history/getDownloadUrl/{agentId}/{sessionId}` | 获取下载链接 | 普通用户 | 生成聊天记录下载链接 |
| GET | `/agent/chat-history/download/{uuid}/current` | 下载当前会话 | 公开 | 下载指定会话的聊天记录（TXT格式） |
| GET | `/agent/chat-history/download/{uuid}/previous` | 下载当前及前20条会话 | 公开 | 下载当前会话及前20条会话记录 |

### 2.4 智能体声纹管理 (`/agent/voice-print`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| POST | `/agent/voice-print` | 创建智能体声纹 | 普通用户 | 为智能体创建声纹 |
| PUT | `/agent/voice-print` | 更新智能体声纹 | 普通用户 | 更新智能体的声纹配置 |
| DELETE | `/agent/voice-print/{id}` | 删除智能体声纹 | 普通用户 | 删除指定的声纹 |
| GET | `/agent/voice-print/list/{id}` | 获取智能体声纹列表 | 普通用户 | 获取指定智能体的所有声纹 |

### 2.5 智能体 MCP 接入点管理 (`/agent/mcp`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/agent/mcp/address/{agentId}` | 获取 MCP 接入点地址 | 普通用户 | 返回智能体的 MCP 接入点地址 |
| GET | `/agent/mcp/tools/{agentId}` | 获取 MCP 工具列表 | 普通用户 | 返回智能体可用的 MCP 工具列表 |

---

## 3. 设备管理模块 (Device)

### 3.1 设备管理 (`/device`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| POST | `/device/bind/{agentId}/{deviceCode}` | 绑定设备 | 普通用户 | 通过验证码绑定设备到智能体 |
| POST | `/device/register` | 注册设备 | 公开 | 生成6位验证码，存储到 Redis |
| GET | `/device/bind/{agentId}` | 获取已绑定设备 | 普通用户 | 返回智能体绑定的所有设备 |
| POST | `/device/bind/{agentId}` | 设备在线状态查询 | 普通用户 | 转发到 MQTT 网关查询设备状态 |
| POST | `/device/unbind` | 解绑设备 | 普通用户 | 解绑设备与智能体的关联 |
| PUT | `/device/update/{id}` | 更新设备信息 | 普通用户 | 更新设备名称等信息 |
| POST | `/device/manual-add` | 手动添加设备 | 普通用户 | 手动添加设备到系统 |

### 3.2 OTA 管理 (`/ota/`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| POST | `/ota/` | OTA版本和设备激活检查 | 公开 | ESP32 设备上报版本和激活状态 |
| POST | `/ota/activate` | 快速检查激活状态 | 公开 | 快速检查设备是否已激活 |
| GET | `/ota/` | OTA接口健康检查 | 公开 | 检查 OTA 接口配置是否正常 |

### 3.3 OTA 固件管理 (`/otaMag`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/otaMag` | 分页查询固件 | 超级管理员 | 分页查询所有 OTA 固件 |
| GET | `/otaMag/{id}` | 获取固件详情 | 超级管理员 | 返回固件详细信息 |
| POST | `/otaMag` | 保存固件信息 | 超级管理员 | 创建新的固件记录 |
| PUT | `/otaMag/{id}` | 修改固件信息 | 超级管理员 | 更新固件信息 |
| DELETE | `/otaMag/{id}` | 删除固件 | 超级管理员 | 删除固件记录 |
| GET | `/otaMag/getDownloadUrl/{id}` | 获取下载链接 | 超级管理员 | 生成固件下载链接 |
| GET | `/otaMag/download/{uuid}` | 下载固件 | 公开 | 通过 UUID 下载固件（限制3次） |
| POST | `/otaMag/upload` | 上传固件 | 超级管理员 | 上传 .bin 或 .apk 文件，使用 MD5 作为文件名 |

---

## 4. 知识库管理模块 (Knowledge)

### 4.1 知识库管理 (`/datasets`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/datasets` | 分页查询知识库 | 普通用户 | 支持按名称查询，只返回当前用户的知识库 |
| GET | `/datasets/{dataset_id}` | 获取知识库详情 | 普通用户 | 返回知识库完整信息 |
| POST | `/datasets` | 创建知识库 | 普通用户 | 创建新的知识库 |
| PUT | `/datasets/{dataset_id}` | 更新知识库 | 普通用户 | 更新知识库信息 |
| DELETE | `/datasets/{dataset_id}` | 删除知识库 | 普通用户 | 删除单个知识库 |
| DELETE | `/datasets/batch` | 批量删除知识库 | 普通用户 | 批量删除多个知识库 |
| GET | `/datasets/rag-models` | 获取 RAG 模型列表 | 普通用户 | 返回可用的 RAG 模型列表 |

### 4.2 知识库文档管理 (`/datasets/{dataset_id}`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/datasets/{dataset_id}/documents` | 分页查询文档 | 普通用户 | 支持按名称和状态查询 |
| GET | `/datasets/{dataset_id}/documents/status/{status}` | 按状态查询文档 | 普通用户 | 根据状态查询文档列表 |
| POST | `/datasets/{dataset_id}/documents` | 上传文档 | 普通用户 | 上传文档到知识库，支持多种格式 |
| DELETE | `/datasets/{dataset_id}/documents/{document_id}` | 删除文档 | 普通用户 | 删除指定文档 |
| POST | `/datasets/{dataset_id}/chunks` | 解析文档（切块） | 普通用户 | 对文档进行切块处理 |
| GET | `/datasets/{dataset_id}/documents/{document_id}/chunks` | 列出文档切片 | 普通用户 | 获取文档的所有切片 |
| POST | `/datasets/{dataset_id}/retrieval-test` | 召回测试 | 普通用户 | 测试知识库的召回效果 |

---

## 5. 模型配置模块 (Model)

### 5.1 模型配置 (`/models`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/models/names` | 获取所有模型名称 | 普通用户 | 根据模型类型获取模型列表 |
| GET | `/models/llm/names` | 获取 LLM 模型信息 | 普通用户 | 返回 LLM 模型列表 |
| GET | `/models/{modelType}/provideTypes` | 获取模型供应器列表 | 超级管理员 | 根据模型类型获取供应器 |
| GET | `/models/list` | 获取模型配置列表 | 超级管理员 | 分页查询模型配置 |
| POST | `/models/{modelType}/{provideCode}` | 新增模型配置 | 超级管理员 | 创建新的模型配置 |
| PUT | `/models/{modelType}/{provideCode}/{id}` | 编辑模型配置 | 超级管理员 | 更新模型配置 |
| DELETE | `/models/{id}` | 删除模型配置 | 超级管理员 | 删除模型配置 |
| GET | `/models/{id}` | 获取模型配置 | 超级管理员 | 返回模型配置详情 |
| PUT | `/models/enable/{id}/{status}` | 启用/关闭模型 | 超级管理员 | 启用或禁用模型配置 |
| PUT | `/models/default/{id}` | 设置默认模型 | 超级管理员 | 设置某个模型为默认模型 |
| GET | `/models/{modelId}/voices` | 获取模型音色 | 普通用户 | 返回模型可用的音色列表 |

### 5.2 模型供应器管理 (`/models/provider`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/models/provider` | 获取模型供应器列表 | 超级管理员 | 分页查询模型供应器 |
| POST | `/models/provider` | 新增模型供应器 | 超级管理员 | 创建新的模型供应器 |
| PUT | `/models/provider` | 修改模型供应器 | 超级管理员 | 更新模型供应器信息 |
| POST | `/models/provider/delete` | 删除模型供应器 | 超级管理员 | 批量删除模型供应器 |
| GET | `/models/provider/plugin/names` | 获取插件名称列表 | 公开 | 返回可用的插件列表 |

---

## 6. 系统管理模块 (Sys)

### 6.1 管理员管理 (`/admin`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/admin/users` | 分页查找用户 | 超级管理员 | 支持按手机号查询 |
| PUT | `/admin/users/{id}` | 重置密码 | 超级管理员 | 重置用户密码并返回新密码 |
| DELETE | `/admin/users/{id}` | 删除用户 | 超级管理员 | 删除指定用户 |
| PUT | `/admin/users/changeStatus/{status}` | 批量修改用户状态 | 超级管理员 | 批量启用/禁用用户 |
| GET | `/admin/device/all` | 分页查找设备 | 超级管理员 | 查询所有用户的设备 |

### 6.2 字典类型管理 (`/admin/dict/type`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/admin/dict/type/page` | 分页查询字典类型 | 超级管理员 | 支持按类型编码和名称查询 |
| GET | `/admin/dict/type/{id}` | 获取字典类型详情 | 超级管理员 | 返回字典类型信息 |
| POST | `/admin/dict/type/save` | 保存字典类型 | 超级管理员 | 创建新的字典类型 |
| PUT | `/admin/dict/type/update` | 修改字典类型 | 超级管理员 | 更新字典类型 |
| POST | `/admin/dict/type/delete` | 删除字典类型 | 超级管理员 | 批量删除字典类型 |

### 6.3 字典数据管理 (`/admin/dict/data`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/admin/dict/data/page` | 分页查询字典数据 | 超级管理员 | 必须指定字典类型ID |
| GET | `/admin/dict/data/{id}` | 获取字典数据详情 | 超级管理员 | 返回字典数据信息 |
| POST | `/admin/dict/data/save` | 新增字典数据 | 超级管理员 | 创建新的字典数据 |
| PUT | `/admin/dict/data/update` | 修改字典数据 | 超级管理员 | 更新字典数据 |
| POST | `/admin/dict/data/delete` | 删除字典数据 | 超级管理员 | 批量删除字典数据 |
| GET | `/admin/dict/data/type/{dictType}` | 获取字典数据列表 | 普通用户 | 根据字典类型获取数据列表 |

### 6.4 参数管理 (`/admin/params`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/admin/params/page` | 分页查询参数 | 超级管理员 | 支持按参数编码查询 |
| GET | `/admin/params/{id}` | 获取参数详情 | 超级管理员 | 返回参数信息 |
| POST | `/admin/params` | 保存参数 | 超级管理员 | 创建新的系统参数 |
| PUT | `/admin/params` | 修改参数 | 超级管理员 | 更新参数，包含多种验证（WebSocket、OTA、MCP、声纹、MQTT密钥） |
| POST | `/admin/params/delete` | 删除参数 | 超级管理员 | 批量删除参数 |

**参数验证规则**:
- WebSocket 地址：不能包含 localhost/127.0.0.1，格式验证，连接测试
- OTA 地址：必须以 http 开头，以 `/ota/` 结尾，连接测试
- MCP 地址：必须包含 "key"，连接测试
- 声纹地址：必须包含 "key"，以 http 开头，健康检查
- MQTT 密钥：长度至少8位，包含大小写字母，不能包含弱密码

### 6.5 服务端管理 (`/admin/server`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/admin/server/server-list` | 获取 WebSocket 服务端列表 | 超级管理员 | 返回所有配置的 WebSocket 地址 |
| POST | `/admin/server/emit-action` | 通知服务端更新配置 | 超级管理员 | 通过 WebSocket 通知 xiaozhi-server 更新配置 |

---

## 7. 音色管理模块 (Timbre)

### 7.1 音色管理 (`/ttsVoice`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/ttsVoice` | 分页查找音色 | 超级管理员 | 必须指定 TTS 模型ID，支持按名称查询 |
| POST | `/ttsVoice` | 保存音色 | 超级管理员 | 创建新的音色配置 |
| PUT | `/ttsVoice/{id}` | 修改音色 | 超级管理员 | 更新音色信息 |
| POST | `/ttsVoice/delete` | 删除音色 | 超级管理员 | 批量删除音色 |

---

## 8. 声音克隆模块 (Voice Clone)

### 8.1 声音克隆管理 (`/voiceClone`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/voiceClone` | 分页查询音色资源 | 普通用户 | 只返回当前用户的音色资源 |
| POST | `/voiceClone/upload` | 上传音频进行声音克隆 | 普通用户 | 上传音频文件（最大10MB） |
| POST | `/voiceClone/updateName` | 更新声音克隆名称 | 普通用户 | 更新克隆声音的名称 |
| POST | `/voiceClone/audio/{id}` | 获取音频下载ID | 普通用户 | 生成音频下载链接 |
| GET | `/voiceClone/play/{uuid}` | 播放音频 | 公开 | 通过 UUID 播放音频 |
| POST | `/voiceClone/cloneAudio` | 复刻音频 | 普通用户 | 开始声音克隆训练 |

### 8.2 音色资源管理 (`/voiceResource`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| GET | `/voiceResource` | 分页查询音色资源 | 超级管理员 | 查询所有用户的音色资源 |
| GET | `/voiceResource/{id}` | 获取音色资源详情 | 超级管理员 | 返回音色资源详细信息 |
| POST | `/voiceResource` | 新增音色资源 | 超级管理员 | 为用户开通音色资源 |
| DELETE | `/voiceResource/{id}` | 删除音色资源 | 超级管理员 | 批量删除音色资源 |
| GET | `/voiceResource/user/{userId}` | 根据用户ID获取音色资源 | 普通用户 | 获取指定用户的音色资源列表 |
| GET | `/voiceResource/ttsPlatforms` | 获取 TTS 平台列表 | 超级管理员 | 返回可用的 TTS 平台列表 |

---

## 9. 配置模块 (Config)

### 9.1 配置获取 (`/config`)

| 方法 | 路径 | 功能 | 权限 | 说明 |
|------|------|------|------|------|
| POST | `/config/server-base` | 服务端获取配置 | 公开 | xiaozhi-server 获取系统配置 |
| POST | `/config/agent-models` | 获取智能体模型 | 公开 | 根据设备 MAC 地址和选择的模块获取模型配置 |

---

## 核心依赖和特性

### 数据库操作
- **ORM**: MyBatis Plus 3.5.5
- **数据库**: MySQL 8.0+
- **连接池**: Druid 1.2.20
- **数据库迁移**: Liquibase 4.20.0

### 缓存
- **Redis**: Spring Data Redis
- **用途**: 
  - 验证码存储
  - Token 存储
  - 临时下载链接
  - 设备验证码

### 安全
- **认证**: Apache Shiro 2.0.2
- **加密**: SM2 国密加密（BouncyCastle）
- **密码加密**: BCrypt
- **权限控制**: 基于 Shiro 的注解权限控制

### 工具库
- **工具类**: Hutool 5.8.24
- **HTTP 客户端**: RestTemplate
- **JSON 处理**: Jackson
- **验证码**: Easy Captcha 1.6.2
- **短信**: 阿里云短信 SDK 4.1.0

### 其他特性
- **API 文档**: Knife4j 4.6.0 (基于 SpringDoc OpenAPI 3)
- **国际化**: 支持中文、英文、德文、越南文、繁体中文
- **XSS 防护**: 内置 XSS 过滤
- **文件上传**: 支持最大 100MB
- **WebSocket**: Spring WebSocket 支持

---

## 数据模型概览

### 主要实体
1. **SysUser**: 系统用户
2. **Agent**: 智能体
3. **AgentTemplate**: 智能体模板
4. **Device**: 设备
5. **KnowledgeBase**: 知识库
6. **KnowledgeFiles**: 知识库文档
7. **ModelConfig**: 模型配置
8. **ModelProvider**: 模型供应器
9. **OtaEntity**: OTA 固件
10. **VoiceClone**: 声音克隆
11. **Timbre**: 音色
12. **SysParams**: 系统参数
13. **SysDictType**: 字典类型
14. **SysDictData**: 字典数据

---

## 迁移到 Golang 的建议

### 1. 框架选择
- **Web 框架**: Gin 或 Echo（推荐 Gin，生态更丰富）
- **ORM**: GORM（功能类似 MyBatis Plus）
- **数据库**: 继续使用 MySQL，使用 `github.com/go-sql-driver/mysql`
- **Redis**: `github.com/redis/go-redis/v9`
- **认证**: JWT（`github.com/golang-jwt/jwt/v5`），可考虑 Casbin 做权限控制

### 2. 项目结构建议
```
manager-api-go/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/          # 配置
│   ├── handler/         # 控制器（对应 Controller）
│   ├── service/         # 服务层
│   ├── repository/      # 数据访问层（对应 Dao）
│   ├── model/           # 数据模型（对应 Entity）
│   ├── dto/             # 数据传输对象
│   ├── middleware/      # 中间件（认证、日志等）
│   └── utils/           # 工具类
├── pkg/                 # 可复用的包
├── api/                 # API 路由定义
└── go.mod
```

### 3. 关键功能实现
- **分页**: 使用 GORM 的 `Scopes` 实现统一分页
- **参数验证**: 使用 `github.com/go-playground/validator/v10`
- **错误处理**: 统一错误码和错误响应格式
- **日志**: 使用 `logrus` 或 `zap`
- **配置管理**: 使用 `viper`
- **API 文档**: 使用 `swaggo/swag` 生成 Swagger 文档

### 4. 需要注意的功能
- **SM2 加密**: 需要找到 Go 的 SM2 实现库
- **文件上传**: Gin 的 `c.SaveUploadedFile` 或 `c.FormFile`
- **WebSocket**: 使用 `gorilla/websocket`
- **短信服务**: 需要集成阿里云短信 Go SDK
- **验证码**: 需要实现图形验证码生成（可使用 `github.com/mojocn/base64Captcha`）

---

## 总结

manager-api 项目包含 **9 个主要模块**，**约 100+ 个 API 端点**，功能涵盖：
- 用户认证与授权
- 智能体全生命周期管理
- 设备管理与 OTA 升级
- 知识库与文档管理
- 模型配置与管理
- 系统参数与字典管理
- 音色与声音克隆
- 配置服务

迁移到 Golang 时，需要重点关注：
1. 保持 API 接口兼容性
2. 数据库表结构保持一致
3. 业务逻辑完全一致
4. 权限控制逻辑一致
5. 错误处理和响应格式一致





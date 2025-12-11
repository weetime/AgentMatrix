# Manager-API Java 到 Go 迁移计划

## 一、项目概览

### 1.1 Java 项目分析 (manager-api)

根据对源码的分析，`manager-api` 主要是一个 **配置下发服务**，核心功能包括：

| 模块 | API 数量 | 主要功能 | 优先级 |
|------|----------|----------|--------|
| **Config 配置模块** | 2 | **核心模块** - 下发配置给 xiaozhi-server | P0 |
| **Device/OTA 设备模块** | 11 | 设备注册、绑定、OTA 升级 | P1 |
| **Agent 智能体模块** | 19 | 智能体 CRUD、聊天记录、声纹、MCP | P2 |
| **Security 认证模块** | 8 | 用户登录、注册、验证码 | P1 |
| **Model 模型模块** | 12 | 模型配置、供应器管理 | P2 |
| **System 系统模块** | 14 | 参数、字典、用户管理 | P3 |
| **Knowledge 知识库** | 8 | 知识库、文档管理 | P3 |
| **Voice 声音模块** | 9 | 音色管理、声音克隆 | P3 |

**总计**: 约 100+ 个 API 端点

### 1.2 Go 项目现状 (manager-api-go)

当前 `manager-api-go` 项目已搭建基础框架：

- ✅ 使用 Kratos v2 框架
- ✅ 使用 Ent ORM（类似 MyBatis Plus）
- ✅ 支持 gRPC + HTTP
- ✅ 支持 MySQL/PostgreSQL/SQLite
- ✅ 已实现 ApiKey 示例模块
- ✅ 支持 OpenTelemetry 链路追踪
- ✅ 支持 Swagger API 文档

### 1.3 技术栈对比

| Java 技术 | Go 替代方案 | 状态 | 说明 |
|-----------|-------------|------|------|
| Spring Boot 3.4.3 | **Kratos v2** | ✅ 已采用 | Web 框架 |
| MyBatis Plus 3.5.5 | **Ent ORM** | ✅ 已采用 | ORM，支持代码生成 |
| Apache Shiro 2.0.2 | **go-jwt + Casbin** | ⏳ 待实现 | JWT 认证 + RBAC 权限 |
| Redis (Spring Data) | **go-redis/v9** | ✅ 已实现 | Redis 客户端，已集成到配置缓存 |
| SM2 国密加密 | **tjfoc/gmsm** | ⏳ 待实现 | 国密 SM2 实现 |
| BCrypt | **golang.org/x/crypto/bcrypt** | ⏳ 待实现 | 密码加密 |
| 图形验证码 (Easy Captcha) | **base64Captcha** | ⏳ 待实现 | 验证码生成 |
| 阿里云短信 SDK | **alibabacloud-go-sdk** | ⏳ 待实现 | 短信服务 |
| Jackson | **encoding/json** | ✅ 内置 | JSON 处理 |
| Knife4j (Swagger) | **swaggo/swag** | ✅ 已配置 | API 文档 |
| Liquibase | **atlas** | ⏳ 待实现 | 数据库迁移 |
| Hutool | **标准库 + 第三方** | ⏳ 部分实现 | 工具类库 |

---

## 二、迁移优先级（基于业务核心程度）

由于 `manager-api` 主要用于 **下发配置**，建议按以下优先级迁移：

### Phase 0: 参数管理下发（P0 - 最高优先级）🔥🔥🔥

**目标**: 实现参数管理的完整功能，包括参数CRUD和配置下发给xiaozhi-server

**当前进度**: 
- ✅ 参数管理 API（0.2）已完成（5个 API）
- ✅ Redis 集成已完成
- ✅ 配置下发 API（0.1）已完成（1个 API）
- ✅ Phase 0 全部完成（6个 API）

#### 0.1 配置下发API（给xiaozhi-server使用）

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 1 | POST | `/config/server-base` | **服务端获取配置** - 返回系统参数（调用buildConfig） | 公开 | 2天 | ✅ 已完成 |

**核心功能**:
- 实现 `buildConfig()` 方法：从 `sys_params` 表读取所有参数，按照 `param_code` 的点号分隔构建嵌套 Map
- 根据 `value_type` 转换数据类型（string, number, boolean, array, json）
- 缓存到 Redis（使用 `RedisKeys.getServerConfigKey()`）
- 支持从缓存读取配置

#### 0.2 参数管理API（管理员使用）

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 2 | GET | `/admin/params/page` | 分页查询参数 | 超级管理员 | 1天 | ✅ 已完成 |
| 3 | GET | `/admin/params/{id}` | 获取参数详情 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 4 | POST | `/admin/params` | 保存参数（含验证） | 超级管理员 | 1.5天 | ✅ 已完成 |
| 5 | PUT | `/admin/params` | 修改参数（含多种验证） | 超级管理员 | 2天 | ✅ 已完成 |
| 6 | POST | `/admin/params/delete` | 删除参数 | 超级管理员 | 0.5天 | ✅ 已完成 |

**参数验证逻辑**:
- WebSocket地址验证（格式、连接测试、禁止localhost）
- OTA地址验证（格式、协议、接口测试）
- MCP地址验证（格式、接口测试）
- 声纹地址验证（格式、接口测试）
- MQTT密钥验证（长度、复杂度、弱密码检测）
- 参数值类型验证（string, number, boolean, array, json）
- 短信参数关联验证

**Phase 0 总计**: 约 7.5 天（1.5周）

**关键依赖**:
- ✅ 已创建 Ent Schema: `sys_params`
- ✅ 已实现 Redis 工具类（`go-redis/v9`）- 位于 `internal/kit/redis.go`，已集成到配置缓存
- ✅ 已实现配置构建逻辑（`buildConfig`）- 位于 `internal/biz/config.go`
- ✅ 已实现参数验证工具类
- ✅ 已实现 WebSocket 连接测试
- ✅ 已实现 HTTP 客户端（用于接口测试）

**实现说明**:
- Redis 客户端封装在 `internal/kit/redis.go`，提供 `GetObject`、`SetObject`、`Delete` 等方法
- 配置缓存使用 `RedisKeyServerConfig = "server:config"` 作为 Key
- 参数管理服务实现在 `internal/service/params.go`，支持完整的 CRUD 操作
- 所有 API 端点已在 `protos/v1/params.proto` 中定义并通过 gRPC Gateway 暴露为 HTTP 接口

---

### Phase 1: 核心配置下发（P0 - 次高优先级）🔥

**目标**: 让 xiaozhi-server 能正常获取完整配置运行（包含模型配置）

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 7 | POST | `/config/agent-models` | **获取智能体模型** - 根据设备MAC返回模型配置 | 公开 | 3天 | ⏳ 进行中 |
| 8 | POST | `/ota/` | OTA版本和设备激活检查 | 公开 | 1天 | ❌ 未开始 |
| 9 | POST | `/ota/activate` | 快速检查激活状态 | 公开 | 0.5天 | ❌ 未开始 |
| 10 | GET | `/ota/` | OTA健康检查 | 公开 | 0.5天 | ❌ 未开始 |

**Phase 1 总计**: 约 5 天（1周）

**关键依赖**:
- 需要创建 Ent Schema: `ai_agent`, `ai_device`, `ai_model_config`, `ai_model_provider`, `ai_agent_template`
- 需要实现配置构建逻辑（`buildModuleConfig`）

---

### Phase 2: 设备管理（P1）

**目标**: 让设备能正常注册和绑定

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 11 | POST | `/device/register` | 设备注册（生成验证码） | 公开 | 1天 | ✅ 已完成 |
| 12 | POST | `/device/bind/{agentId}/{deviceCode}` | 设备绑定 | 普通用户 | 1天 | ✅ 已完成 |
| 13 | GET | `/device/bind/{agentId}` | 获取已绑定设备 | 普通用户 | 0.5天 | ✅ 已完成 |
| 14 | POST | `/device/bind/{agentId}` | 设备在线状态查询（转发MQTT） | 普通用户 | 1天 | ✅ 已完成 |
| 15 | POST | `/device/unbind` | 解绑设备 | 普通用户 | 0.5天 | ✅ 已完成 |
| 16 | PUT | `/device/update/{id}` | 更新设备信息 | 普通用户 | 0.5天 | ✅ 已完成 |
| 17 | POST | `/device/manual-add` | 手动添加设备 | 普通用户 | 0.5天 | ✅ 已完成 |

**Phase 2 总计**: 约 5 天（1周）

**关键依赖**:
- 需要创建 Ent Schema: `ai_device`
- 需要实现 MQTT 网关转发逻辑
- 需要实现 Bearer Token 生成（SHA256）

---

### Phase 3: 认证模块（P1）

**目标**: 管理员和用户登录功能

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 18 | GET | `/user/captcha` | 获取图形验证码 | 公开 | 1天 | ✅ 已完成 |
| 19 | POST | `/user/smsVerification` | 发送短信验证码 | 公开 | 1.5天 | ✅ 已完成 |
| 20 | POST | `/user/login` | 用户登录（支持SM2加密） | 公开 | 2天 | ✅ 已完成 |
| 21 | POST | `/user/register` | 用户注册 | 公开 | 1.5天 | ✅ 已完成 |
| 22 | GET | `/user/info` | 获取当前用户信息 | 普通用户 | 0.5天 | ✅ 已完成 |
| 23 | PUT | `/user/change-password` | 修改密码 | 普通用户 | 1天 | ✅ 已完成 |
| 24 | PUT | `/user/retrieve-password` | 找回密码 | 公开 | 1天 | ✅ 已完成 |
| 25 | GET | `/user/pub-config` | 获取公共配置 | 公开 | 0.5天 | ✅ 已完成 |

**Phase 3 总计**: 约 8 天（1.5周）

**关键依赖**:
- 需要创建 Ent Schema: `sys_user`, `sys_user_token`
- 需要实现 SM2 加密/解密
- 需要实现图形验证码生成
- 需要集成阿里云短信 SDK
- 需要实现 JWT Token 生成和验证
- 需要实现认证中间件

---

### Phase 4: 智能体管理（P2）

**目标**: 智能体 CRUD 操作

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 26 | GET | `/agent/list` | 获取用户智能体列表 | 普通用户 | 0.5天 | ✅ 已完成 |
| 27 | GET | `/agent/all` | 智能体列表（管理员） | 超级管理员 | 0.5天 | ✅ 已完成 |
| 28 | GET | `/agent/{id}` | 获取智能体详情 | 普通用户 | 0.5天 | ✅ 已完成 |
| 29 | POST | `/agent` | 创建智能体 | 普通用户 | 1天 | ✅ 已完成 |
| 30 | PUT | `/agent/saveMemory/{macAddress}` | 根据设备更新智能体记忆 | 公开 | 0.5天 | ✅ 已完成 |
| 31 | PUT | `/agent/{id}` | 更新智能体 | 普通用户 | 1天 | ✅ 已完成 |
| 32 | DELETE | `/agent/{id}` | 删除智能体（级联删除） | 普通用户 | 1天 | ✅ 已完成 |
| 33 | GET | `/agent/template` | 获取智能体模板列表 | 普通用户 | 0.5天 | ✅ 已完成 |
| 34 | GET | `/agent/{id}/sessions` | 获取智能体会话列表 | 普通用户 | 0.5天 | ✅ 已完成 |
| 35 | GET | `/agent/{id}/chat-history/{sessionId}` | 获取智能体聊天记录 | 普通用户 | 0.5天 | ✅ 已完成 |
| 36 | GET | `/agent/{id}/chat-history/user` | 获取智能体最近50条聊天记录 | 普通用户 | 0.5天 | ✅ 已完成 |
| 37 | GET | `/agent/{id}/chat-history/audio` | 获取音频内容 | 普通用户 | 0.5天 | ✅ 已完成 |
| 38 | POST | `/agent/audio/{audioId}` | 获取音频下载ID | 普通用户 | 0.5天 | ✅ 已完成 |
| 39 | GET | `/agent/play/{uuid}` | 播放音频 | 公开 | 0.5天 | ✅ 已完成 |

**智能体模板管理**:
| 40 | GET | `/agent/template/page` | 分页查询模板 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 41 | GET | `/agent/template/{id}` | 获取模板详情 | 超级管理员 | 0.5天 | ✅ 已完成|
| 42 | POST | `/agent/template` | 创建模板 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 43 | PUT | `/agent/template` | 更新模板 | 超级管理员 | 0.5天 | ✅ 已完成|
| 44 | DELETE | `/agent/template/{id}` | 删除模板 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 45 | POST | `/agent/template/batch-remove` | 批量删除模板 | 超级管理员 | 0.5天 | ✅ 已完成 |

**聊天历史管理**:
| 46 | POST | `/agent/chat-history/report` | 小智服务聊天上报 | 公开 | 1天 | ❌ 未开始 |
| 47 | POST | `/agent/chat-history/getDownloadUrl/{agentId}/{sessionId}` | 获取下载链接 | 普通用户 | 0.5天 | ❌ 未开始 |
| 48 | GET | `/agent/chat-history/download/{uuid}/current` | 下载当前会话 | 公开 | 0.5天 | ❌ 未开始 |
| 49 | GET | `/agent/chat-history/download/{uuid}/previous` | 下载当前及前20条会话 | 公开 | 0.5天 | ❌ 未开始 |

**声纹管理**:
| 50 | POST | `/agent/voice-print` | 创建智能体声纹 | 普通用户 | 1天 | ✅ 已完成 |
| 51 | PUT | `/agent/voice-print` | 更新智能体声纹 | 普通用户 | 0.5天 | ✅ 已完成 |
| 52 | DELETE | `/agent/voice-print/{id}` | 删除智能体声纹 | 普通用户 | 0.5天 | ✅ 已完成 |
| 53 | GET | `/agent/voice-print/list/{id}` | 获取智能体声纹列表 | 普通用户 | 0.5天 | ✅ 已完成 |

**MCP 接入点管理**:
| 54 | GET | `/agent/mcp/address/{agentId}` | 获取 MCP 接入点地址 | 普通用户 | 0.5天 | ✅ 已完成  |
| 55 | GET | `/agent/mcp/tools/{agentId}` | 获取 MCP 工具列表 | 普通用户 | 0.5天 | ✅ 已完成  |

**Phase 4 总计**: 约 20 天（4周）

**关键依赖**:
- 需要创建 Ent Schema: `ai_agent`, `ai_agent_template`, `ai_agent_chat_history`, `ai_agent_chat_audio`, `ai_agent_plugin_mapping`, `ai_agent_voice_print`
- 需要实现文件上传/下载功能
- 需要实现级联删除逻辑

---

### Phase 5: 模型配置（P2）

**目标**: 模型配置和供应器管理

| 序号 | 方法 | 路径 | 功能 | 权限 | 预计工时 | 状态 |
|------|------|------|------|------|----------|------|
| 56 | GET | `/models/names` | 获取所有模型名称 | 普通用户 | 0.5天 | ✅ 已实现 |
| 57 | GET | `/models/llm/names` | 获取 LLM 模型信息 | 普通用户 | 0.5天 | ✅ 已实现 |
| 58 | GET | `/models/{modelType}/provideTypes` | 获取模型供应器列表 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 59 | GET | `/models/list` | 获取模型配置列表 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 60 | POST | `/models/{modelType}/{provideCode}` | 新增模型配置 | 超级管理员 | 1天 | ✅ 已实现 |
| 61 | PUT | `/models/{modelType}/{provideCode}/{id}` | 编辑模型配置 | 超级管理员 | 1天 | ✅ 已实现 |
| 62 | DELETE | `/models/{id}` | 删除模型配置 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 63 | GET | `/models/{id}` | 获取模型配置 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 64 | PUT | `/models/enable/{id}/{status}` | 启用/关闭模型 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 65 | PUT | `/models/default/{id}` | 设置默认模型 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 66 | GET | `/models/{modelId}/voices` | 获取模型音色 | 普通用户 | 0.5天 | ⚠️ 部分实现（返回空列表，需完善） |

**模型供应器管理**:
| 67 | GET | `/models/provider` | 获取模型供应器列表 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 68 | POST | `/models/provider` | 新增模型供应器 | 超级管理员 | 1天 | ✅ 已实现 |
| 69 | PUT | `/models/provider` | 修改模型供应器 | 超级管理员 | 1天 | ✅ 已实现 |
| 70 | POST | `/models/provider/delete` | 删除模型供应器 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 71 | GET | `/models/provider/plugin/names` | 获取插件名称列表 | 公开 | 0.5天 | ✅ 已实现（有TODO，基本功能已实现） |

**音色管理**:
| 72 | GET | `/ttsVoice` | 分页查找音色 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 73 | POST | `/ttsVoice` | 保存音色 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 74 | PUT | `/ttsVoice/{id}` | 修改音色 | 超级管理员 | 0.5天 | ✅ 已实现 |
| 75 | POST | `/ttsVoice/delete` | 删除音色 | 超级管理员 | 0.5天 | ✅ 已实现 |

**Phase 5 总计**: 约 12 天（2.5周）

**关键依赖**:
- 需要创建 Ent Schema: `ai_model_config`, `ai_model_provider`, `ai_tts_voice`
- 需要实现模型配置的 JSON 解析和验证

---

### Phase 6: 系统管理（P3）

**目标**: 系统参数、字典、用户管理

**管理员管理**:
| 76 | GET | `/admin/users` | 分页查找用户 | 超级管理员 | 0.5天 | ✅ 已完成 
| 77 | PUT | `/admin/users/{id}` | 重置密码 | 超级管理员 | 0.5天 | ✅ 已完成 
| 78 | DELETE | `/admin/users/{id}` | 删除用户 | 超级管理员 | 0.5天 | ✅ 已完成 
| 79 | PUT | `/admin/users/changeStatus/{status}` | 批量修改用户状态 | 超级管理员 | 0.5 天 | ✅ 已完成 
| 80 | GET | `/admin/device/all` | 分页查找设备 | 超级管理员 | 0.5天 | ✅ 已废弃不用实现 

**字典类型管理**:
| 81 | GET | `/admin/dict/type/page` | 分页查询字典类型 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 82 | GET | `/admin/dict/type/{id}` | 获取字典类型详情 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 83 | POST | `/admin/dict/type/save` | 保存字典类型 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 84 | PUT | `/admin/dict/type/update` | 修改字典类型 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 85 | POST | `/admin/dict/type/delete` | 删除字典类型 | 超级管理员 | 0.5天 | ✅ 已完成 |

**字典数据管理**:
| 86 | GET | `/admin/dict/data/page` | 分页查询字典数据 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 87 | GET | `/admin/dict/data/{id}` | 获取字典数据详情 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 88 | POST | `/admin/dict/data/save` | 新增字典数据 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 89 | PUT | `/admin/dict/data/update` | 修改字典数据 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 90 | POST | `/admin/dict/data/delete` | 删除字典数据 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 91 | GET | `/admin/dict/data/type/{dictType}` | 获取字典数据列表 | 普通用户 | 0.5天 | ✅ 已完成 |

**服务端管理**:
| 92 | GET | `/admin/server/server-list` | 获取 WebSocket 服务端列表 | 超级管理员 | 0.5天 | ⚠️待测试
| 93 | POST | `/admin/server/emit-action` | 通知服务端更新配置 | 超级管理员 | 1天 | ⚠️待测试

**OTA 固件管理**:
| 94 | GET | `/otaMag` | 分页查询固件 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 95 | GET | `/otaMag/{id}` | 获取固件详情 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 96 | POST | `/otaMag` | 保存固件信息 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 97 | PUT | `/otaMag/{id}` | 修改固件信息 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 98 | DELETE | `/otaMag/{id}` | 删除固件 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 99 | GET | `/otaMag/getDownloadUrl/{id}` | 获取下载链接 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 100 | GET | `/otaMag/download/{uuid}` | 下载固件 | 公开 | 0.5天 | ✅ 已完成 |
| 101 | POST | `/otaMag/upload` | 上传固件 | 超级管理员 | 1天 | ✅ 已完成 |

**Phase 6 总计**: 约 15 天（3周）

**关键依赖**:
- 需要创建 Ent Schema: `sys_dict_type`, `sys_dict_data`, `ai_ota`
- 需要实现参数验证逻辑（WebSocket、OTA、MCP、声纹、MQTT密钥）
- 需要实现 WebSocket 通知功能
- 需要实现文件上传/下载功能

---

### Phase 7: 高级功能（P3）

**知识库管理**:
| 102 | GET | `/datasets` | 分页查询知识库 | 普通用户 | 0.5天 | ✅ 已完成 |
| 103 | GET | `/datasets/{dataset_id}` | 获取知识库详情 | 普通用户 | 0.5天 | ✅ 已完成 |
| 104 | POST | `/datasets` | 创建知识库 | 普通用户 | 0.5天 | ✅ 已完成 |
| 105 | PUT | `/datasets/{dataset_id}` | 更新知识库 | 普通用户 | 0.5天 | ✅ 已完成 |
| 106 | DELETE | `/datasets/{dataset_id}` | 删除知识库 | 普通用户 | 0.5天 | ✅ 已完成 |
| 107 | DELETE | `/datasets/batch` | 批量删除知识库 | 普通用户 | 0.5天 | ✅ 已完成 |
| 108 | GET | `/datasets/rag-models` | 获取 RAG 模型列表 | 普通用户 | 0.5天 | ✅ 已完成 |

**知识库文档管理**:
| 109 | GET | `/datasets/{dataset_id}/documents` | 分页查询文档 | 普通用户 | 0.5天 | ✅ 已完成 |
| 110 | GET | `/datasets/{dataset_id}/documents/status/{status}` | 按状态查询文档 | 普通用户 | 0.5天 | ✅ 已完成 |
| 111 | POST | `/datasets/{dataset_id}/documents` | 上传文档 | 普通用户 | 1天 | ✅ 已完成 |
| 112 | DELETE | `/datasets/{dataset_id}/documents/{document_id}` | 删除文档 | 普通用户 | 0.5天 | ✅ 已完成 |
| 113 | POST | `/datasets/{dataset_id}/chunks` | 解析文档（切块） | 普通用户 | 1天 | ✅ 已完成 |
| 114 | GET | `/datasets/{dataset_id}/documents/{document_id}/chunks` | 列出文档切片 | 普通用户 | 0.5天 | ✅ 已完成 |
| 115 | POST | `/datasets/{dataset_id}/retrieval-test` | 召回测试 | 普通用户 | 1天 | ✅ 已完成 |

**声音克隆管理**:
| 116 | GET | `/voiceClone` | 分页查询音色资源 | 普通用户 | 0.5天 | ✅ 已完成 |
| 117 | POST | `/voiceClone/upload` | 上传音频进行声音克隆 | 普通用户 | 1天 | ✅ 已完成 |
| 118 | POST | `/voiceClone/updateName` | 更新声音克隆名称 | 普通用户 | 0.5天 | ✅ 已完成 |
| 119 | POST | `/voiceClone/audio/{id}` | 获取音频下载ID | 普通用户 | 0.5天 | ✅ 已完成 |
| 120 | GET | `/voiceClone/play/{uuid}` | 播放音频 | 公开 | 0.5天 | ✅ 已完成 |
| 121 | POST | `/voiceClone/cloneAudio` | 复刻音频 | 普通用户 | 1天 | ✅ 已完成 |

**音色资源管理**:
| 122 | GET | `/voiceResource` | 分页查询音色资源 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 123 | GET | `/voiceResource/{id}` | 获取音色资源详情 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 124 | POST | `/voiceResource` | 新增音色资源 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 125 | DELETE | `/voiceResource/{id}` | 删除音色资源 | 超级管理员 | 0.5天 | ✅ 已完成 |
| 126 | GET | `/voiceResource/user/{userId}` | 根据用户ID获取音色资源 | 普通用户 | 0.5天 | ✅ 已完成 |
| 127 | GET | `/voiceResource/ttsPlatforms` | 获取 TTS 平台列表 | 超级管理员 | 0.5天 | ✅ 已完成 |

**Phase 7 总计**: 约 12 天（2.5周）

**关键依赖**:
- 需要创建 Ent Schema: `ai_rag_dataset`, `ai_voice_clone`
- 需要实现文件上传/下载功能
- 需要实现文档解析和切块功能

---

## 三、数据库 Schema 迁移

### 3.1 需要创建的 Ent Schema

需要创建以下 Ent Schema（对应 Java 的 Entity）:

```
internal/data/ent/schema/
├── sys_user.go                  # 系统用户 ✅ Phase 3
├── sys_user_token.go            # 用户 Token ✅ Phase 3
├── sys_params.go                # 系统参数 ✅ Phase 1
├── sys_dict_type.go             # 字典类型 ✅ Phase 6
├── sys_dict_data.go             # 字典数据 ✅ Phase 6
├── ai_agent.go                  # 智能体 ✅ Phase 1, 4
├── ai_agent_template.go         # 智能体模板 ✅ Phase 1, 4
├── ai_device.go                 # 设备 ✅ Phase 1, 2
├── ai_model_config.go           # 模型配置 ✅ Phase 1, 5
├── ai_model_provider.go         # 模型供应器 ✅ Phase 1, 5
├── ai_tts_voice.go              # 音色 ✅ Phase 5
├── ai_ota.go                    # OTA 固件 ✅ Phase 6
├── ai_agent_chat_history.go     # 聊天记录 ✅ Phase 4
├── ai_agent_chat_audio.go       # 聊天音频 ✅ Phase 4
├── ai_agent_plugin_mapping.go   # 插件映射 ✅ Phase 1, 4
├── ai_agent_voice_print.go      # 声纹 ✅ Phase 4
├── ai_voice_clone.go            # 声音克隆 ✅ Phase 7
├── ai_rag_dataset.go            # 知识库 ✅ Phase 7
└── api_key.go                   # ✅ 已存在
```

### 3.2 Schema 创建顺序

1. **Phase 0 所需** (优先级最高):
   - `sys_params.go` ✅ 参数管理核心表

2. **Phase 1 所需** (次高优先级):
   - `ai_agent.go`
   - `ai_agent_template.go`
   - `ai_device.go`
   - `ai_model_config.go`
   - `ai_model_provider.go`
   - `ai_agent_plugin_mapping.go`

2. **Phase 2 所需**:
   - 无新增（使用 Phase 1 的 `ai_device.go`）

3. **Phase 3 所需**:
   - `sys_user.go`
   - `sys_user_token.go`

4. **Phase 4 所需**:
   - `ai_agent_chat_history.go`
   - `ai_agent_chat_audio.go`
   - `ai_agent_voice_print.go`

5. **Phase 5 所需**:
   - `ai_tts_voice.go`

6. **Phase 6 所需**:
   - `sys_dict_type.go`
   - `sys_dict_data.go`
   - `ai_ota.go`

7. **Phase 7 所需**:
   - `ai_voice_clone.go`
   - `ai_rag_dataset.go`

---

## 四、项目目录结构

```
manager-api-go/
├── cmd/
│   └── server/
│       └── main.go                    # 应用入口
├── config/
│   └── config.dev.yaml                # 配置文件
├── internal/
│   ├── app.go                         # Wire 依赖注入
│   ├── biz/                           # 业务逻辑层（UseCase）
│   │   ├── config.go                  # Phase 0: 配置业务逻辑（buildConfig）
│   │   ├── params.go                  # Phase 0: 参数管理业务逻辑
│   │   ├── device.go                  # Phase 2: 设备业务逻辑
│   │   ├── ota.go                     # Phase 1: OTA 业务逻辑
│   │   ├── user.go                    # Phase 3: 用户业务逻辑
│   │   ├── agent.go                   # Phase 4: 智能体业务逻辑
│   │   ├── model.go                   # Phase 5: 模型业务逻辑
│   │   ├── sys.go                     # Phase 6: 系统管理业务逻辑
│   │   ├── knowledge.go               # Phase 7: 知识库业务逻辑
│   │   ├── voice.go                   # Phase 7: 声音克隆业务逻辑
│   │   └── biz.go                     # Wire Provider
│   ├── data/                          # 数据访问层（Repository）
│   │   ├── ent/                       # Ent 生成的代码
│   │   │   ├── schema/                # Schema 定义
│   │   │   │   ├── sys_user.go
│   │   │   │   ├── sys_params.go     # ✅ Phase 0: 参数管理表
│   │   │   │   ├── ai_agent.go
│   │   │   │   └── ...
│   │   │   └── ...
│   │   ├── config.go                  # Phase 0: 配置数据访问
│   │   ├── params.go                  # Phase 0: 参数管理数据访问
│   │   ├── device.go                  # Phase 2: 设备数据访问
│   │   ├── user.go                    # Phase 3: 用户数据访问
│   │   ├── agent.go                   # Phase 4: 智能体数据访问
│   │   ├── model.go                   # Phase 5: 模型数据访问
│   │   ├── sys.go                     # Phase 6: 系统数据访问
│   │   ├── knowledge.go               # Phase 7: 知识库数据访问
│   │   ├── voice.go                   # Phase 7: 声音数据访问
│   │   ├── data.go                    # Wire Provider
│   │   └── page_helper.go             # 分页辅助函数
│   ├── service/                       # API 服务层（对应 Controller）
│   │   ├── config.go                  # Phase 0: 配置服务（/config/server-base）
│   │   ├── params.go                  # Phase 0: 参数管理服务（/admin/params）
│   │   ├── device.go                  # Phase 2: 设备服务
│   │   ├── ota.go                     # Phase 1: OTA 服务
│   │   ├── user.go                    # Phase 3: 用户服务
│   │   ├── agent.go                   # Phase 4: 智能体服务
│   │   ├── model.go                   # Phase 5: 模型服务
│   │   ├── sys.go                     # Phase 6: 系统服务
│   │   ├── knowledge.go               # Phase 7: 知识库服务
│   │   ├── voice.go                   # Phase 7: 声音服务
│   │   ├── service.go                 # Wire Provider
│   │   └── transform.go              # 数据转换辅助函数
│   ├── server/                        # HTTP/gRPC 服务
│   │   ├── http.go                    # HTTP 路由注册
│   │   ├── grpc.go                    # gRPC 服务注册
│   │   └── server.go                  # Wire Provider
│   └── kit/                           # 工具类库
│       ├── redis.go                   # ✅ Phase 0: Redis 工具类
│       ├── validator.go               # ✅ Phase 0: 参数验证工具（WebSocket、OTA、MCP等）
│       ├── sm2.go                     # SM2 加密工具
│       ├── captcha.go                 # 验证码工具
│       ├── jwt.go                     # JWT 工具
│       ├── bcrypt.go                  # BCrypt 密码加密
│       ├── websocket.go               # ✅ Phase 0: WebSocket 连接测试工具
│       ├── mqtt.go                    # MQTT 网关工具
│       ├── sms.go                     # 短信服务工具
│       ├── page_request.go            # 分页请求处理
│       ├── error.go                   # 错误处理
│       ├── log.go                     # 日志工具
│       └── validate.go                # 参数验证
├── protos/                            # Proto 定义
│   └── agent-matrix/v1/
│       ├── config.proto               # ✅ Phase 0: 配置服务
│       ├── params.proto               # ✅ Phase 0: 参数管理服务
│       ├── device.proto               # Phase 2: 设备服务
│       ├── ota.proto                  # Phase 1: OTA 服务
│       ├── user.proto                 # Phase 3: 用户服务
│       ├── agent.proto                # Phase 4: 智能体服务
│       ├── model.proto                # Phase 5: 模型服务
│       ├── sys.proto                  # Phase 6: 系统服务
│       ├── knowledge.proto            # Phase 7: 知识库服务
│       ├── voice.proto                # Phase 7: 声音服务
│       └── error.proto                # 错误定义
├── docs/                              # Swagger 文档
│   └── swagger.json
├── go.mod
├── go.sum
├── Makefile
├── API_FEATURES.md                    # API 功能清单
└── MIGRATION_PLAN.md                  # 本文档
```

---

## 五、技术实现细节

### 5.1 关键依赖清单

需要在 `go.mod` 中添加的依赖：

```go
// Redis
github.com/redis/go-redis/v9 v9.x.x

// 国密 SM2
github.com/tjfoc/gmsm v1.4.1

// 验证码
github.com/mojocn/base64Captcha v1.3.6

// JWT
github.com/golang-jwt/jwt/v5 v5.2.1

// BCrypt
golang.org/x/crypto v0.x.x

// 阿里云短信
github.com/alibabacloud-go/dysmsapi-20170525/v4 v4.x.x
github.com/alibabacloud-go/tea v1.x.x

// HTTP 客户端（用于 MQTT 网关转发）
github.com/go-resty/resty/v2 v2.x.x

// WebSocket 客户端
github.com/gorilla/websocket v1.5.3  // ✅ 已存在

// 文件上传
github.com/gin-gonic/gin v1.x.x  // 或使用 Kratos 内置的文件处理

// UUID
github.com/google/uuid v1.6.0  // ✅ 已存在

// 数据拷贝
github.com/jinzhu/copier v0.4.0  // ✅ 已存在
```

### 5.2 核心功能实现要点

#### 5.2.1 配置下发逻辑 (`buildConfig`) - Phase 0 核心功能

**Java 实现位置**: `ConfigServiceImpl.buildConfig()`

**Go 实现要点**:
- 从 `sys_params` 表查询所有参数（`param_type = 1` 非系统参数）
- 按照 `param_code` 的点号分隔构建嵌套 Map（如 `server.ip` -> `{"server": {"ip": "0.0.0.0"}}`）
- 根据 `value_type` 转换数据类型：
  - `string`: 直接使用字符串值
  - `number`: 解析为数字（整数或浮点数）
  - `boolean`: 解析为布尔值
  - `array`: 按分号分隔转换为字符串数组
  - `json`: 解析为 JSON 对象
- 缓存到 Redis（使用 `RedisKeys.getServerConfigKey()`）
- 支持从缓存读取配置（`isCache = true`）

**关键代码结构**:
```go
func (uc *ConfigUsecase) buildConfig(ctx context.Context) (map[string]interface{}, error) {
    // 1. 查询所有系统参数（param_type = 1）
    // 2. 遍历参数，按照 param_code 点号分隔构建嵌套 Map
    // 3. 根据 value_type 转换数据类型
    // 4. 返回构建好的配置 Map
}

func (uc *ConfigUsecase) GetConfig(ctx context.Context, isCache bool) (map[string]interface{}, error) {
    // 1. 如果 isCache = true，先从 Redis 读取
    // 2. 如果缓存不存在，调用 buildConfig() 构建配置
    // 3. 将配置存入 Redis
    // 4. 返回配置
}
```

#### 5.2.2 模型配置构建 (`buildModuleConfig`)

**Java 实现位置**: `ConfigServiceImpl.buildModuleConfig()`

**Go 实现要点**:
- 根据模型类型（VAD, ASR, TTS, LLM, Memory, Intent, VLLM, RAG）查询模型配置
- 处理 TTS 音色注入（`private_voice`, `ref_audio`, `ref_text`）
- 处理 Intent 模型的附加 LLM 模型
- 处理 Memory 模型的附加 LLM 模型
- 构建 `selected_module` Map

**关键代码结构**:
```go
func (uc *ConfigUsecase) buildModuleConfig(
    ctx context.Context,
    agent *biz.Agent,
    result map[string]interface{},
) error {
    // 1. 遍历模型类型
    // 2. 查询模型配置
    // 3. 处理特殊逻辑（TTS、Intent、Memory）
    // 4. 构建 selected_module
}
```

#### 5.2.3 SM2 加密/解密

**Java 实现位置**: `SM2Utils`, `Sm2DecryptUtil`

**Go 实现要点**:
- 使用 `github.com/tjfoc/gmsm` 库
- 实现 SM2 公钥加密
- 实现 SM2 私钥解密
- 处理密钥格式转换（Base64）

**关键代码结构**:
```go
package kit

import (
    "github.com/tjfoc/gmsm/sm2"
)

func SM2Encrypt(publicKey string, plaintext string) (string, error) {
    // SM2 加密实现
}

func SM2Decrypt(privateKey string, ciphertext string) (string, error) {
    // SM2 解密实现
}
```

#### 5.2.4 JWT Token 认证

**Java 实现位置**: `SysUserTokenService`, Shiro 配置

**Go 实现要点**:
- 使用 `github.com/golang-jwt/jwt/v5`
- 实现 Token 生成（包含用户ID、过期时间）
- 实现 Token 验证中间件
- 存储 Token 到 Redis（可选，用于 Token 黑名单）

**关键代码结构**:
```go
package kit

import (
    "github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID int64, username string) (string, error) {
    // JWT Token 生成
}

func ValidateToken(tokenString string) (*Claims, error) {
    // JWT Token 验证
}
```

#### 5.2.5 图形验证码

**Java 实现位置**: `CaptchaController`

**Go 实现要点**:
- 使用 `github.com/mojocn/base64Captcha`
- 生成数字验证码
- 存储验证码到 Redis（使用 UUID 作为 key）
- 返回 Base64 图片和 UUID

**关键代码结构**:
```go
package kit

import (
    "github.com/mojocn/base64Captcha"
)

func GenerateCaptcha() (string, string, error) {
    // 生成验证码图片和 UUID
}
```

#### 5.2.6 MQTT 网关转发

**Java 实现位置**: `DeviceController.forwardToMqttGateway()`

**Go 实现要点**:
- 从系统参数获取 MQTT 网关地址
- 生成 Bearer Token（SHA256(date + signature_key)）
- 构建设备 ID 列表（格式：`groupId@@@macAddress@@@macAddress`）
- 转发 POST 请求到 MQTT 网关

**关键代码结构**:
```go
package kit

func ForwardToMqttGateway(ctx context.Context, deviceIds []string) (string, error) {
    // 1. 获取 MQTT 网关地址
    // 2. 生成 Bearer Token
    // 3. 构建请求
    // 4. 发送 HTTP POST 请求
}
```

---

## 六、实施计划（时间表）

### 第零阶段：参数管理下发（1.5周）🔥🔥🔥

**Week 1**:
- [x] Day 1: 创建 `sys_params.go` Ent Schema
  - 定义表结构（id, param_code, param_value, value_type, param_type, remark等）
  - 生成 Ent 代码
- [x] Day 2: 实现 Redis 工具类
  - 配置 Redis 连接
  - 实现基本的 get/set/delete 操作
  - 实现参数缓存逻辑
- [x] Day 3-4: 实现 `buildConfig()` 核心逻辑
  - 从数据库查询所有系统参数
  - 按照 `param_code` 点号分隔构建嵌套 Map
  - 根据 `value_type` 转换数据类型（string, number, boolean, array, json）
  - 实现 Redis 缓存逻辑
- [x] Day 5: 实现 `POST /config/server-base` API
  - 实现配置服务接口
  - 支持从缓存读取
  - 编写单元测试

**Week 2**:
- [x] Day 1: 实现参数管理基础 API
  - `GET /admin/params/page` - 分页查询
  - `GET /admin/params/{id}` - 获取详情
- [x] Day 2: 实现参数保存和删除 API
  - `POST /admin/params` - 保存参数（含基础验证）
  - `POST /admin/params/delete` - 删除参数
- [x] Day 3-4: 实现参数修改 API（含复杂验证）
  - `PUT /admin/params` - 修改参数
  - 实现 WebSocket 地址验证
  - 实现 OTA 地址验证
  - 实现 MCP 地址验证
  - 实现声纹地址验证
  - 实现 MQTT 密钥验证
  - 实现参数值类型验证
- [x] Day 5: 集成测试和文档
  - 编写集成测试
  - 验证配置下发功能
  - 更新 API 文档

**交付物**: 
- ✅ Phase 0 的 6 个 API 完全可用
- ✅ 参数管理功能完整（CRUD + 验证）
- ✅ 配置下发功能可用（xiaozhi-server 可以获取系统参数配置）
- ✅ Redis 缓存正常工作

---

### 第一阶段：核心配置下发（1周）🔥

**Week 3**:
- [ ] Day 1-2: 创建 Phase 1 所需的 Ent Schema
  - `ai_agent.go`
  - `ai_agent_template.go`
  - `ai_device.go`
  - `ai_model_config.go`
  - `ai_model_provider.go`
  - `ai_agent_plugin_mapping.go`
- [ ] Day 3-4: 实现 `buildModuleConfig()` 逻辑
  - 实现模型配置查询和构建
  - 实现 TTS 音色注入
  - 实现 Intent/Memory 附加模型处理
- [ ] Day 5: 实现 `POST /config/agent-models` API
  - 根据设备MAC查询智能体
  - 构建完整的模型配置
  - 实现 OTA 相关 API（3个）

**交付物**: 
- ✅ Phase 1 的 4 个 API 完全可用
- ✅ xiaozhi-server 可以正常获取完整配置并运行

---

### 第二阶段：设备管理 + 认证（2周）

**Week 4**:
- [ ] Day 1-2: 实现设备管理 API（6个）
  - 设备注册、绑定、解绑、更新、手动添加
  - MQTT 网关转发
- [x] Day 3-4: 实现 JWT 认证中间件
  - Token 生成和验证
  - 认证中间件
- [x] Day 5: 实现用户基础 API（获取用户信息）

**Week 5**:
- [x] Day 1: 实现图形验证码
- [x] Day 2-3: 实现用户登录（支持 SM2 加密）
- [x] Day 4: 实现用户注册
- [x] Day 5: 实现密码修改和找回密码

**交付物**:
- ✅ Phase 2 的 6 个设备管理 API
- ✅ Phase 3 的 8 个认证 API
- ✅ 管理员和用户可以正常登录

---

### 第三阶段：智能体 + 模型管理（3周）

**Week 6-7**:
- [x] 创建 Phase 4 所需的 Ent Schema
  - `ai_agent_chat_history.go`
  - `ai_agent_chat_audio.go`
  - `ai_agent_voice_print.go`
- [x] 实现智能体 CRUD API（14个基础 API）
- [ ] 实现智能体模板管理 API（6个）
- [x] 实现聊天记录 API（4个基础 API）
- [ ] 实现声纹管理 API（4个）
- [ ] 实现 MCP 接入点 API（2个）

**Week 7**:
- [ ] 创建 Phase 5 所需的 Ent Schema
  - `ai_tts_voice.go`
- [ ] 实现模型配置 API（11个）
- [ ] 实现模型供应器 API（5个）
- [ ] 实现音色管理 API（4个）

**交付物**:
- ✅ Phase 4 的 30 个智能体相关 API
- ✅ Phase 5 的 20 个模型相关 API

---

### 第四阶段：系统管理（2周）

**Week 9**:
- [x] 创建 Phase 6 所需的 Ent Schema
  - `sys_dict_type.go`
  - `sys_dict_data.go`
  - `ai_ota.go`
- [ ] 实现管理员管理 API（5个）
- [x] 实现字典管理 API（11个）
- [x] 实现参数管理 API（5个）

**Week 10**:
- [ ] 实现服务端管理 API（2个）
- [ ] 实现 OTA 固件管理 API（8个）
- [x] 实现参数验证逻辑（WebSocket、OTA、MCP、声纹、MQTT密钥）
- [ ] 实现 WebSocket 通知功能

**交付物**:
- ✅ Phase 6 的 31 个系统管理 API

---

### 第五阶段：高级功能（可选，2周）

**Week 11-12**:
- [ ] 创建 Phase 7 所需的 Ent Schema
  - `ai_voice_clone.go`
  - `ai_rag_dataset.go`
- ✅ 实现知识库管理 API（7个）
- ✅ 实现知识库文档管理 API（7个）
- [ ] 实现声音克隆管理 API（6个）
- [ ] 实现音色资源管理 API（6个）

**交付物**:
- ✅ Phase 7 的知识库管理 API（14个已完成）
- ⏳ Phase 7 的声音克隆和音色资源管理 API（12个待实现）

---

## 七、关键注意事项

### 7.1 API 兼容性

- ✅ **路径保持一致**: 所有 API 路径必须与 Java 版本完全一致（包括 `/xiaozhi` 前缀）
- ✅ **请求格式一致**: 请求参数格式、Content-Type 保持一致
- ✅ **响应格式一致**: 响应 JSON 结构、字段名称保持一致
- ✅ **错误码一致**: 错误码和错误消息保持一致

### 7.2 数据库兼容性

- ✅ **表结构一致**: Ent Schema 必须与 Java 版本的数据库表结构完全一致
- ✅ **字段类型一致**: 字段类型、长度、约束保持一致
- ✅ **索引一致**: 索引定义保持一致

### 7.3 业务逻辑一致性

- ✅ **配置构建逻辑**: `buildConfig()` 和 `buildModuleConfig()` 的逻辑必须完全一致
- ✅ **权限控制**: 权限验证逻辑保持一致
- ✅ **数据验证**: 参数验证规则保持一致

### 7.4 性能优化

- ✅ **配置缓存**: 系统配置需要缓存到 Redis，减少数据库查询
- ✅ **分页查询**: 所有列表查询必须支持分页
- ✅ **数据库连接池**: 合理配置数据库连接池大小

### 7.5 安全性

- ✅ **密码加密**: 使用 BCrypt 加密存储密码
- ✅ **SM2 加密**: 登录密码使用 SM2 加密传输
- ✅ **Token 安全**: JWT Token 设置合理的过期时间
- ✅ **SQL 注入防护**: 使用 Ent ORM 的参数化查询
- ✅ **XSS 防护**: 对用户输入进行 XSS 过滤

---

## 八、测试策略

### 8.1 单元测试

- 每个 `biz` 层的 UseCase 都需要编写单元测试
- 每个工具类都需要编写单元测试
- 测试覆盖率目标：**80%+**

### 8.2 集成测试

- 每个 Phase 完成后，编写集成测试
- 使用实际的 xiaozhi-server 进行端到端测试
- 验证 API 响应格式与 Java 版本一致

### 8.3 性能测试

- 配置下发 API 的响应时间 < 100ms
- 列表查询 API 的响应时间 < 500ms
- 支持并发请求数 > 1000 QPS

---

## 九、风险评估与应对

### 9.1 技术风险

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| Ent Schema 与数据库表结构不一致 | 高 | 仔细对比 Java Entity 和数据库表结构 |
| 配置构建逻辑复杂，容易出错 | 高 | 编写详细的单元测试，逐行对比 Java 代码 |
| SM2 加密实现差异 | 中 | 使用成熟的国密库，编写加密/解密测试 |
| 性能不如 Java 版本 | 中 | 优化数据库查询，使用 Redis 缓存 |

### 9.2 业务风险

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| API 不兼容导致 xiaozhi-server 无法使用 | 高 | 每个 Phase 完成后立即进行集成测试 |
| 数据迁移问题 | 中 | 使用相同的数据库，无需数据迁移 |
| 权限控制不一致 | 高 | 仔细对比 Shiro 权限配置和 Go 实现 |

---

## 十、总结

### 10.1 迁移统计

| 项目 | 数量 | 已完成 | 进行中 | 未开始 |
|------|------|--------|--------|--------|
| **总 API 数量** | ~127 个 | **39 个** | **1 个** | **87 个** |
| **Phase 0 (P0)** | 6 个（参数管理 + 配置下发） | ✅ **6 个** | - | - |
| **Phase 1 (P0)** | 4 个（完整配置下发 + OTA） | - | ⏳ **1 个** | ❌ **3 个** |
| **Phase 2-3 (P1)** | 15 个（设备管理 + 认证） | ✅ **8 个** | - | ❌ **7 个** |
| **Phase 4-5 (P2)** | 50 个（智能体 + 模型） | ✅ **14 个** | - | ❌ **36 个** |
| **Phase 6-7 (P3)** | 52 个（系统管理 + 高级功能） | ✅ **11 个** | - | ❌ **41 个** |
| **预计总工期** | 10-12 周（不含 Phase 7 则 8-10 周） | - | - | - |
| **Phase 0 交付时间** | 1.5 周内可让参数管理和配置下发正常工作 | ✅ **已完成** | - | - |

### 10.2 关键里程碑

- ✅ **Week 2**: Phase 0 完成，参数管理和配置下发功能可用
- ⏳ **Week 3**: Phase 1 进行中，xiaozhi-server 可以正常获取完整配置（部分完成）
- ✅ **Week 5**: Phase 3 完成，用户认证功能可用（8个 API）
- ⏳ **Week 6-7**: Phase 4 部分完成，智能体基础功能可用（14个 API）
- ⏳ **Week 9**: Phase 6 部分完成，字典和参数管理功能可用（11个 API）
- ❌ **Week 8**: Phase 4-5 待完成，智能体和模型管理完整功能
- ❌ **Week 10**: Phase 6 待完成，系统管理功能完整
- ❌ **Week 12**: Phase 7 待完成（可选），所有功能迁移完成

### 10.3 下一步行动

1. **Phase 0 进度**:
   - ✅ 已创建 `sys_params.go` Ent Schema
   - ✅ 已实现 Redis 工具类（`internal/kit/redis.go`）
   - ✅ 已实现 `buildConfig()` 核心逻辑（`internal/biz/config.go`）
   - ✅ 已实现参数管理 CRUD API（`internal/service/params.go`）
   - ✅ 已实现配置下发 API (`POST /config/server-base`)
   - ✅ 确保与 xiaozhi-server 兼容

2. **Phase 3 进度**:
   - ✅ 已实现认证模块所有 API（8个）
   - ✅ 已实现用户登录、注册、密码管理等功能
   - ✅ 已实现图形验证码和短信验证码

3. **Phase 4 进度**:
   - ✅ 已实现智能体基础 CRUD API（14个）
   - ✅ 已实现智能体聊天记录相关 API（4个）
   - ⏳ 待实现智能体模板管理 API（6个）
   - ⏳ 待实现声纹管理 API（4个）
   - ⏳ 待实现 MCP 接入点 API（2个）

4. **Phase 6 进度**:
   - ✅ 已实现字典类型管理 API（5个）
   - ✅ 已实现字典数据管理 API（6个）
   - ✅ 已实现参数管理 API（5个）
   - ⏳ 待实现管理员管理 API（5个）
   - ⏳ 待实现服务端管理 API（2个）
   - ⏳ 待实现 OTA 固件管理 API（8个）

5. **Phase 1 进度**:
   - ⏳ 正在实现配置下发 API (`POST /config/agent-models`)
   - ❌ 待实现 OTA 相关 API（3个）

6. **并行工作**:
   - 前端团队可以开始准备 Go 版本的 API 调用
   - 测试团队可以准备测试用例

7. **持续集成**:
   - 每个 Phase 完成后立即部署到测试环境
   - 进行集成测试和性能测试

---

## 附录

### A. 参考文档

- [Kratos 官方文档](https://go-kratos.dev/)
- [Ent 官方文档](https://entgo.io/)
- [API_FEATURES.md](./API_FEATURES.md) - API 功能清单

### B. 联系方式

如有问题，请参考：
- Java 源码：`main/manager-api/src/main/java/`
- Go 源码：`main/manager-api-go/internal/`
- 数据库迁移脚本：`main/manager-api/src/main/resources/db/changelog/`

---

**文档版本**: v1.0  
**最后更新**: 2025-01-XX  
**维护者**: 开发团队



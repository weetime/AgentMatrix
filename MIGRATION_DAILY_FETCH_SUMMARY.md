# Daily Fetch PR 迁移总结

## 迁移完成时间
2025-12-16

## 迁移内容概述

本次迁移将 daily-fetch PR 中的 Java manager-api 变更迁移到 Go AgentMatrix 项目，主要包括：

1. **AgentContextProvider 功能** - 智能体上下文源配置管理
2. **WebSocket Token 认证** - 设备注册时生成 WebSocket 认证 Token
3. **系统参数更新** - 新增 `server.auth.enabled` 和 `system-web.menu` 参数
4. **设备注册优化** - 验证码生成逻辑优化（避免重复）
5. **前端配置支持** - GetPubConfig 返回 systemWebMenu 配置

---

## 已完成的迁移任务

### ✅ Task 1: 创建 AgentContextProvider 数据库 Schema
- **文件**: `internal/data/ent/schema/agent_context_provider.go`
- **状态**: 已完成
- **说明**: 创建了 `ai_agent_context_provider` 表的 Ent Schema，包含 id, agent_id, context_providers (JSON), creator, created_at, updater, updated_at 字段

### ✅ Task 2: 创建 AgentContextProvider 业务层
- **文件**: `internal/biz/agent_context_provider.go`
- **状态**: 已完成
- **说明**: 实现了 `AgentContextProviderUsecase`，包含 GetByAgentId、SaveOrUpdateByAgentId、DeleteByAgentId 方法

### ✅ Task 3: 创建 AgentContextProvider 数据访问层
- **文件**: `internal/data/agent_context_provider.go`
- **状态**: 已完成
- **说明**: 实现了 `AgentContextProviderRepo`，使用 Ent 进行数据库操作，处理 JSON 字段的序列化/反序列化

### ✅ Task 4: 更新 Agent Service 层
- **文件**: `internal/service/agent.go`
- **状态**: 已完成
- **说明**: 
  - 在 `GetAgentById` 中查询并返回 contextProviders
  - 在 `UpdateAgent` 中处理 contextProviders 的保存/更新
  - 在 `DeleteAgent` 中调用删除上下文源配置

### ✅ Task 5: 更新 Agent Proto 定义
- **文件**: `protos/v1/agent.proto`
- **状态**: 已完成
- **说明**: 
  - 新增 `ContextProvider` message
  - 在 `AgentInfoVO` 中增加 `context_providers` 字段
  - 在 `AgentUpdateRequest` 中增加 `context_providers` 字段

### ✅ Task 6: 更新 Config Service 层
- **文件**: `internal/biz/config.go`
- **状态**: 已完成
- **说明**: 在 `buildAgentConfig` 方法中增加上下文源配置的构建逻辑，在 MCP 接入点之后、声纹信息之前添加

### ✅ Task 7: 实现 WebSocket Token 生成功能
- **文件**: `internal/kit/crypto.go`, `internal/biz/ota.go`
- **状态**: 已完成
- **说明**: 
  - 在 `kit` 包中新增 `GenerateWebSocketToken` 函数
  - 在 `CheckDeviceActive` 方法中根据 `server.auth.enabled` 参数决定是否生成 token
  - Token 格式：`signature.timestamp`（HMAC-SHA256 + Base64 URL-safe）

### ✅ Task 8: 优化设备注册验证码生成
- **文件**: `internal/service/device.go`
- **状态**: 已完成（已存在）
- **说明**: 验证码生成逻辑已优化，使用循环检查 Redis 中是否已存在该验证码

### ✅ Task 9: 更新常量定义
- **文件**: `internal/constant/constant.go`, `internal/service/user.go`
- **状态**: 已完成
- **说明**: 
  - 新增 `ServerAuthEnabled` 常量：`"server.auth.enabled"`
  - 更新版本号为 `"0.8.10"`

### ✅ Task 10: 更新 User Service 层
- **文件**: `internal/service/user.go`
- **状态**: 已完成
- **说明**: 在 `GetPubConfig` 方法中增加 `systemWebMenu` 配置返回，从系统参数中读取 `system-web.menu` 参数（JSON 格式）

### ✅ Task 11: 更新系统参数初始化
- **文件**: `migrations/20251216_daily_fetch_params.sql`
- **状态**: 已完成
- **说明**: 创建了数据库迁移脚本，包含：
  - `server.auth.enabled` 参数（id=122）
  - `system-web.menu` 参数（id=600）
  - `ai_agent_context_provider` 表创建

### ✅ Task 12: 更新 Wire 依赖注入
- **文件**: `internal/biz/biz.go`, `internal/data/data.go`
- **状态**: 已完成
- **说明**: 
  - 在 `biz.go` 中注册 `NewAgentContextProviderUsecase`
  - 在 `data.go` 中注册 `NewAgentContextProviderRepo`
  - 更新 `NewConfigUsecase` 和 `NewAgentService` 的构造函数签名

---

## 新增文件清单

1. `internal/data/ent/schema/agent_context_provider.go` - Ent Schema 定义
2. `internal/biz/agent_context_provider.go` - 业务逻辑层
3. `internal/data/agent_context_provider.go` - 数据访问层
4. `migrations/20251216_daily_fetch_params.sql` - 数据库迁移脚本
5. `MIGRATION_DAILY_FETCH_SUMMARY.md` - 本文档

---

## 修改文件清单

1. `protos/v1/agent.proto` - 新增 ContextProvider message 和相关字段
2. `internal/service/agent.go` - 更新 GetAgentById、UpdateAgent、DeleteAgent 方法
3. `internal/biz/config.go` - 更新 ConfigUsecase 结构体和 buildAgentConfig 方法
4. `internal/biz/ota.go` - 更新 CheckDeviceActive 方法，添加 WebSocket Token 生成
5. `internal/kit/crypto.go` - 新增 GenerateWebSocketToken 函数
6. `internal/service/user.go` - 更新 GetPubConfig 方法，添加 systemWebMenu 配置
7. `internal/constant/constant.go` - 新增 ServerAuthEnabled 常量和 Version
8. `internal/biz/biz.go` - 注册 NewAgentContextProviderUsecase
9. `internal/data/data.go` - 注册 NewAgentContextProviderRepo

---

## 数据库变更

### 新增表
- `ai_agent_context_provider` - 智能体上下文源配置表

### 新增系统参数
- `server.auth.enabled` (id=122, type='boolean', value='true') - server模块是否开启token认证
- `system-web.menu` (id=600, type='json') - 系统功能菜单配置

---

## 后续步骤

1. **执行数据库迁移**：
   ```bash
   mysql -u username -p database_name < migrations/20251216_daily_fetch_params.sql
   ```

2. **重新生成代码**：
   ```bash
   cd main/AgentMatrix
   make generate_ent
   make generate_proto
   make generate_wire
   ```

3. **编译测试**：
   ```bash
   go build ./...
   ```

4. **功能测试**：
   - 测试 AgentContextProvider CRUD 功能
   - 测试智能体更新时上下文源配置的保存/更新
   - 测试配置下发时上下文源配置的包含
   - 测试 WebSocket Token 生成（与 Python 端对比）
   - 测试 systemWebMenu 配置返回

---

## 注意事项

1. **JSON 字段处理**: `context_providers` 字段是 JSON 类型，Ent 会将其作为字符串存储和读取，需要手动进行 JSON 序列化/反序列化

2. **Token 生成算法**: WebSocket Token 生成必须与 Python 端保持一致（HMAC-SHA256 + Base64 URL-safe），格式为 `signature.timestamp`

3. **系统参数**: 确保新增的系统参数在数据库中存在，可以通过执行迁移脚本或手动插入

4. **API 兼容性**: 确保 Proto 定义的字段名称与 Java 版本保持一致（使用 snake_case）

5. **依赖注入**: Wire 代码需要重新生成以确保依赖注入正确

---

## 参考文档

- [迁移计划](./MIGRATION_PLAN.md)
- [API 迁移分析](./API_MIGRATION_ANALYSIS.md)
- Java 源码参考：`main/manager-api/src/main/java/xiaozhi/modules/agent/`

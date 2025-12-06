# MCP 接入点工作原理和流程分析

## 一、WebSocket 文件说明

### 1. `websocket.go` - 服务器端实现（**不应删除**）
**作用**：作为 WebSocket 服务器，接收来自 Node 服务（xiaozhi-server）的连接

**主要功能**：
- 处理来自 Node 服务的 WebSocket 连接请求（`NodeWebSocketHandler`）
- 管理节点连接池（`nodeConnections`）
- 心跳检测和连接管理
- 广播消息给节点服务

**使用场景**：
- 当 xiaozhi-server 需要连接到 AgentMatrix 时使用
- 用于设备管理和消息推送

### 2. `websocket_client.go` - 客户端实现（**新添加**）
**作用**：作为 WebSocket 客户端，连接到 MCP 服务器

**主要功能**：
- 连接到外部 MCP 服务器
- 发送 JSON-RPC 消息
- 监听和过滤响应消息
- 管理连接生命周期

**使用场景**：
- 获取 MCP 工具列表时使用
- 需要与外部 MCP 服务器通信时使用

**结论**：两个文件作用不同，`websocket.go` **不应删除**。

---

## 二、MCP 接入点完整工作流程

### 架构图

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│  前端/客户端     │         │  AgentMatrix      │         │  MCP 服务器      │
│  (Manager API)  │────────▶│  (Go 服务)        │────────▶│  (xiaozhi-server)│
└─────────────────┘         └──────────────────┘         └─────────────────┘
                                      │
                                      │ WebSocket Client
                                      │
                                      ▼
                            ┌──────────────────┐
                            │  MCP Endpoint    │
                            │  (外部 MCP 服务) │
                            └──────────────────┘
```

### 流程 1：获取 MCP 接入点地址

**API**: `GET /agent/mcp/address/{agentId}`

**步骤**：

1. **权限检查**
   - 从 Context 获取当前用户ID
   - 检查用户是否有权限访问该智能体
   - 超级管理员可以访问所有智能体

2. **获取配置**
   ```
   从系统参数表获取: server.mcp_endpoint
   例如: https://mcp.example.com:8080/path?key=your_secret_key
   ```

3. **解析 URL**
   - 解析 MCP 端点配置 URL
   - 提取协议、主机、路径、查询参数

4. **生成 WebSocket URL**
   ```
   原始 URL: https://mcp.example.com:8080/path?key=your_secret_key
   
   步骤：
   a. 协议转换: https → wss, http → ws
   b. 提取路径前缀: /path → /path (去掉最后一个/)
   c. 提取密钥: key=your_secret_key → your_secret_key
   
   结果: wss://mcp.example.com:8080/path
   ```

5. **生成 Token**
   ```
   步骤：
   a. 对 agentId 进行 MD5 哈希: MD5(agentId) → "abc123..."
   b. 构建 JSON: {"agentId": "abc123..."}
   c. 使用 AES/ECB/PKCS5Padding 加密: AES(key, json) → "encrypted_token"
   d. URL 编码: URLEncode("encrypted_token") → "encrypted_token%3D..."
   ```

6. **构建最终地址**
   ```
   最终地址: wss://mcp.example.com:8080/path/mcp/?token=encrypted_token%3D...
   ```

7. **返回结果**
   ```json
   {
     "code": 0,
     "msg": "success",
     "data": "wss://mcp.example.com:8080/path/mcp/?token=..."
   }
   ```

### 流程 2：获取 MCP 工具列表

**API**: `GET /agent/mcp/tools/{agentId}`

**步骤**：

1. **权限检查**（同流程1）

2. **获取 MCP 地址**
   - 调用 `GetAgentMcpAccessAddress` 获取地址
   - 如果地址为空，返回空列表

3. **转换 URL**
   ```
   将 /mcp/ 替换为 /call/
   原地址: wss://.../mcp/?token=...
   新地址: wss://.../call/?token=...
   ```

4. **建立 WebSocket 连接**
   ```
   配置：
   - 连接超时: 8 秒
   - 最大会话时长: 10 秒
   - 缓冲区大小: 1MB
   ```

5. **MCP 初始化流程（JSON-RPC 2.0）**

   **步骤 5.1: 发送初始化请求**
   ```json
   {
     "jsonrpc": "2.0",
     "method": "initialize",
     "params": {
       "protocolVersion": "2024-11-05",
       "capabilities": {
         "roots": {"listChanged": false},
         "sampling": {}
       },
       "clientInfo": {
         "name": "xz-mcp-broker",
         "version": "0.0.1"
       }
     },
     "id": 1
   }
   ```

   **步骤 5.2: 等待初始化响应**
   ```json
   {
     "jsonrpc": "2.0",
     "id": 1,
     "result": {
       "protocolVersion": "2024-11-05",
       "capabilities": {...},
       "serverInfo": {...}
     }
   }
   ```
   - 监听 id=1 的响应
   - 检查是否有 `result` 字段且无 `error` 字段
   - 如果失败，返回空列表

   **步骤 5.3: 发送初始化完成通知**
   ```json
   {
     "jsonrpc": "2.0",
     "method": "notifications/initialized"
   }
   ```

   **步骤 5.4: 发送工具列表请求**
   ```json
   {
     "jsonrpc": "2.0",
     "method": "tools/list",
     "params": null,
     "id": 2
   }
   ```

   **步骤 5.5: 等待工具列表响应**
   ```json
   {
     "jsonrpc": "2.0",
     "id": 2,
     "result": {
       "tools": [
         {
           "name": "tool1",
           "description": "...",
           "inputSchema": {...}
         },
         {
           "name": "tool2",
           ...
         }
       ]
     }
   }
   ```
   - 监听 id=2 的响应
   - 提取 `result.tools` 数组
   - 提取每个工具的 `name` 字段

6. **返回工具列表**
   ```json
   {
     "code": 0,
     "msg": "success",
     "data": ["tool1", "tool2", "tool3"]
   }
   ```

7. **关闭连接**
   - 自动关闭 WebSocket 连接
   - 清理资源

---

## 三、安全机制

### 1. Token 生成流程
```
agentId → MD5 → JSON → AES加密 → URL编码 → Token
```

### 2. 权限控制
- 用户只能访问自己创建的智能体
- 超级管理员可以访问所有智能体
- Token 中包含智能体ID的加密信息

### 3. 连接安全
- 使用 WSS（WebSocket Secure）协议
- Token 通过 URL 参数传递
- 连接有超时限制（10秒）

---

## 四、错误处理

### 1. MCP 地址未配置
```json
{
  "code": 0,
  "msg": "success",
  "data": "请联系管理员进入参数管理配置mcp接入点地址"
}
```

### 2. 权限不足
```json
{
  "code": 403,
  "msg": "没有权限查看该智能体的MCP接入点地址/工具列表"
}
```

### 3. WebSocket 连接失败
- 返回空列表 `[]`
- 记录警告日志

### 4. MCP 初始化失败
- 返回空列表 `[]`
- 记录错误日志

---

## 五、关键代码位置

### AgentMatrix (Go)
- **Protobuf 定义**: `protos/v1/agent.proto`
- **Service 层**: `internal/service/agent.go` (GetAgentMcpAddress, GetAgentMcpTools)
- **Biz 层**: `internal/biz/agent.go` (GetAgentMcpAccessAddress, GetAgentMcpToolsList)
- **加密工具**: `internal/kit/crypto.go` (AESEncrypt, MD5HexDigest)
- **MCP JSON-RPC**: `internal/kit/mcp_jsonrpc.go` (GetInitializeJson, GetToolsListJson)
- **WebSocket 客户端**: `internal/kit/websocket_client.go`

### xiaozhi-server (Python)
- **MCP 端点处理器**: `core/providers/tools/mcp_endpoint/mcp_endpoint_handler.py`
- **MCP 客户端**: `core/providers/tools/mcp_endpoint/mcp_endpoint_client.py`

---

## 六、配置说明

### 系统参数配置
```
参数编码: server.mcp_endpoint
参数值格式: https://mcp.example.com:8080/path?key=your_secret_key
```

### URL 格式要求
- 必须包含 `key` 查询参数（AES 加密密钥）
- 支持 http/https 协议（会自动转换为 ws/wss）
- 路径部分会被用于构建 WebSocket URL

---

## 七、使用示例

### 1. 获取 MCP 地址
```bash
curl -X GET "http://localhost:8080/agent/mcp/address/agent123" \
  -H "Authorization: Bearer your_token"
```

### 2. 获取工具列表
```bash
curl -X GET "http://localhost:8080/agent/mcp/tools/agent123" \
  -H "Authorization: Bearer your_token"
```

---

## 八、总结

MCP 接入点系统实现了：
1. **安全的 Token 生成机制**：通过 MD5 + AES 加密保护智能体ID
2. **标准的 JSON-RPC 2.0 协议**：与 MCP 服务器通信
3. **完整的错误处理**：各种异常情况的处理
4. **权限控制**：确保用户只能访问自己的智能体
5. **WebSocket 客户端**：用于与外部 MCP 服务器通信

两个 WebSocket 文件的作用：
- `websocket.go`: 服务器端，接收 Node 服务连接
- `websocket_client.go`: 客户端，连接外部 MCP 服务器

**两者功能不同，都需要保留。**


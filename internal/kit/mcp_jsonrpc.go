package kit

import (
	"encoding/json"
)

// JsonRpcTwo JSON-RPC 2.0 请求结构
type JsonRpcTwo struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

// GetInitializeJson 获取MCP初始化JSON
func GetInitializeJson() string {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": false,
			},
			"sampling": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "xz-mcp-broker",
			"version": "0.0.1",
		},
	}
	
	req := JsonRpcTwo{
		JsonRPC: "2.0",
		Method:  "initialize",
		Params:  params,
		ID:      1,
	}
	
	jsonBytes, _ := json.Marshal(req)
	return string(jsonBytes)
}

// GetNotificationsInitializedJson 获取MCP初始化完成通知JSON
func GetNotificationsInitializedJson() string {
	return `{"jsonrpc":"2.0","method":"notifications/initialized"}`
}

// GetToolsListJson 获取MCP工具列表请求JSON
func GetToolsListJson() string {
	req := JsonRpcTwo{
		JsonRPC: "2.0",
		Method:  "tools/list",
		Params:  nil,
		ID:      2,
	}
	
	jsonBytes, _ := json.Marshal(req)
	return string(jsonBytes)
}


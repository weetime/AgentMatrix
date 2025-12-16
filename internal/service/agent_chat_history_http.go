package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/constant"
	"github.com/weetime/agent-matrix/internal/middleware"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// RegisterAgentChatHistoryHTTPHandlers 注册聊天历史管理的HTTP handlers
func RegisterAgentChatHistoryHTTPHandlers(srv *kratoshttp.Server, agentService *AgentService) {
	// 注意：静态路由必须在动态路由之前注册，防止路由被覆盖
	// 1. 静态路由：下载当前会话
	srv.HandlePrefix("/agent/chat-history/download/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/current") {
			agentService.DownloadCurrentSessionHandler(w, r)
		} else if strings.HasSuffix(path, "/previous") {
			agentService.DownloadPreviousSessionsHandler(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	// 2. 静态路由：聊天上报
	srv.HandleFunc("/agent/chat-history/report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		agentService.ReportChatHistoryHandler(w, r)
	})

	// 3. 动态路由：获取下载链接（需要认证）
	srv.HandleFunc("/agent/chat-history/getDownloadUrl/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		agentService.GetDownloadUrlHandler(w, r)
	})
}

// ReportChatHistoryHandler 处理聊天上报请求
func (s *AgentService) ReportChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 解析请求体
	var req struct {
		MacAddress  string  `json:"macAddress"`
		SessionID   string  `json:"sessionId"`
		ChatType    int8    `json:"chatType"`
		Content     string  `json:"content"`
		AudioBase64 *string `json:"audioBase64,omitempty"`
		ReportTime  *int64  `json:"reportTime,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, 400, "请求参数解析失败: "+err.Error())
		return
	}

	// 验证必填字段
	if req.MacAddress == "" || req.SessionID == "" || req.Content == "" {
		writeErrorResponse(w, http.StatusBadRequest, 400, "macAddress、sessionId和content不能为空")
		return
	}
	if req.ChatType != 1 && req.ChatType != 2 {
		writeErrorResponse(w, http.StatusBadRequest, 400, "chatType必须为1（用户）或2（智能体）")
		return
	}

	// 构建业务请求
	bizReq := &biz.ReportChatHistoryRequest{
		MacAddress:  req.MacAddress,
		SessionID:   req.SessionID,
		ChatType:    req.ChatType,
		Content:     req.Content,
		AudioBase64: req.AudioBase64,
		ReportTime:  req.ReportTime,
	}

	// 调用业务逻辑
	result, err := s.uc.ReportChatHistory(ctx, bizReq)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, 500, err.Error())
		return
	}

	// 返回响应
	writeSuccessResponse(w, result)
}

// GetDownloadUrlHandler 处理获取下载链接请求
func (s *AgentService) GetDownloadUrlHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从路径中提取 agentId 和 sessionId
	// 路径格式: /agent/chat-history/getDownloadUrl/{agentId}/{sessionId}
	path := r.URL.Path
	pathPrefix := "/agent/chat-history/getDownloadUrl/"
	if !strings.HasPrefix(path, pathPrefix) {
		writeErrorResponse(w, http.StatusBadRequest, 400, "无效的路径")
		return
	}

	// 移除前缀，获取剩余部分
	remaining := strings.TrimPrefix(path, pathPrefix)
	parts := strings.Split(remaining, "/")
	if len(parts) != 2 {
		writeErrorResponse(w, http.StatusBadRequest, 400, "路径参数格式错误，应为 /agent/chat-history/getDownloadUrl/{agentId}/{sessionId}")
		return
	}

	agentId := parts[0]
	sessionId := parts[1]

	if agentId == "" || sessionId == "" {
		writeErrorResponse(w, http.StatusBadRequest, 400, "agentId和sessionId不能为空")
		return
	}

	// 获取当前用户信息
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, 401, "未授权，请先登录")
		return
	}

	// 检查权限
	hasPermission, err := s.uc.CheckAgentPermission(ctx, agentId, user.ID, user.SuperAdmin == 1)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, 500, err.Error())
		return
	}
	if !hasPermission {
		writeErrorResponse(w, http.StatusForbidden, constant.ErrorCodeChatHistoryNoPermission, "没有权限查看该智能体的聊天记录")
		return
	}

	// 生成下载链接
	uuid, err := s.uc.GetChatHistoryDownloadUrl(ctx, agentId, sessionId)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, 500, err.Error())
		return
	}

	// 返回响应
	writeSuccessResponse(w, uuid)
}

// DownloadCurrentSessionHandler 处理下载当前会话请求
func (s *AgentService) DownloadCurrentSessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从路径中提取 uuid
	// 路径格式: /agent/chat-history/download/{uuid}/current
	path := r.URL.Path
	pathPrefix := "/agent/chat-history/download/"
	if !strings.HasPrefix(path, pathPrefix) {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}

	remaining := strings.TrimPrefix(path, pathPrefix)
	if !strings.HasSuffix(remaining, "/current") {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}

	uuid := strings.TrimSuffix(remaining, "/current")
	if uuid == "" {
		http.Error(w, "uuid不能为空", http.StatusBadRequest)
		return
	}

	// 调用业务逻辑下载聊天记录
	content, err := s.uc.DownloadChatHistory(ctx, uuid, false)
	if err != nil {
		if strings.Contains(err.Error(), "下载链接已过期") || strings.Contains(err.Error(), "下载链接无效") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("下载失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
	fileName := url.QueryEscape("history.txt")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

	// 写入响应流
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(content)
	if err != nil {
		// 写入失败，但响应头已发送，无法返回错误
		return
	}
}

// DownloadPreviousSessionsHandler 处理下载当前及前20条会话请求
func (s *AgentService) DownloadPreviousSessionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从路径中提取 uuid
	// 路径格式: /agent/chat-history/download/{uuid}/previous
	path := r.URL.Path
	pathPrefix := "/agent/chat-history/download/"
	if !strings.HasPrefix(path, pathPrefix) {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}

	remaining := strings.TrimPrefix(path, pathPrefix)
	if !strings.HasSuffix(remaining, "/previous") {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}

	uuid := strings.TrimSuffix(remaining, "/previous")
	if uuid == "" {
		http.Error(w, "uuid不能为空", http.StatusBadRequest)
		return
	}

	// 调用业务逻辑下载聊天记录（包含前20条会话）
	content, err := s.uc.DownloadChatHistory(ctx, uuid, true)
	if err != nil {
		if strings.Contains(err.Error(), "下载链接已过期") || strings.Contains(err.Error(), "下载链接无效") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("下载失败: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
	fileName := url.QueryEscape("history.txt")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

	// 写入响应流
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(content)
	if err != nil {
		// 写入失败，但响应头已发送，无法返回错误
		return
	}
}

// writeSuccessResponse 写入成功响应（与Java的Result<T>格式一致）
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	// Java的Result<T>格式: {"code": 0, "msg": "success", "data": <T>}
	response := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": data,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, 500, "序列化响应失败: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// writeErrorResponse 写入错误响应（与Java的Result<T>格式一致）
func writeErrorResponse(w http.ResponseWriter, httpStatus int, code int32, msg string) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	// Java的Result<T>格式: {"code": <code>, "msg": "<msg>", "data": null}
	response := map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "序列化错误响应失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(httpStatus)
	w.Write(jsonData)
}

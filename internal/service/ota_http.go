package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/constant"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// handleCORS 处理CORS头和OPTIONS预检请求
// 如果返回true，表示已经处理了OPTIONS请求，调用者应该直接返回
func handleCORS(w http.ResponseWriter, r *http.Request, allowedMethods string, allowedHeaders string) bool {
	// 设置CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
	w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")

	// 处理OPTIONS预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

// DownloadOtaHandler 处理固件下载的HTTP handler
func (s *OtaService) DownloadOtaHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从路径中提取uuid
	// 路径格式: /otaMag/download/{uuid}
	path := r.URL.Path
	pathPrefix := "/otaMag/download/"
	if !strings.HasPrefix(path, pathPrefix) {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}
	uuidStr := strings.TrimPrefix(path, pathPrefix)
	if uuidStr == "" {
		http.Error(w, "uuid不能为空", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 调用biz层下载固件
	fileData, filename, err := s.uc.DownloadOta(ctx, uuidStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("下载失败: %v", err), http.StatusNotFound)
		return
	}
	if fileData == nil || len(fileData) == 0 {
		http.Error(w, "固件文件不存在", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 写入文件数据
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(fileData)
	if err != nil {
		// 写入失败，但响应头已发送，无法返回错误
		return
	}
}

// UploadFirmwareHandler 处理固件上传的HTTP handler
func (s *OtaService) UploadFirmwareHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 权限检查：超级管理员
	ctx := r.Context()

	// 从HTTP请求中提取Token（因为直接注册的handler可能不经过中间件）
	token := extractTokenFromRequest(r)
	if token == "" {
		response := &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证Token并获取用户信息
	user, err := s.tokenService.GetUserByToken(ctx, token)
	if err != nil || user == nil {
		response := &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 检查账号状态
	if user.Status == 0 {
		response := &pb.Response{
			Code: 401,
			Msg:  "账号已被锁定",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 将用户信息存储到Context中
	ctx = context.WithValue(ctx, middleware.UserDetailKey, user)

	// 检查是否为超级管理员
	if user.SuperAdmin != 1 {
		response := &pb.Response{
			Code: 403,
			Msg:  "需要超级管理员权限",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 解析multipart/form-data
	err = r.ParseMultipartForm(constant.MaxUploadSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
		return
	}

	// 获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("获取文件失败: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件类型
	originalFilename := header.Filename
	if originalFilename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext != ".bin" && ext != ".apk" {
		http.Error(w, "只允许上传.bin和.apk格式的文件", http.StatusBadRequest)
		return
	}

	// 验证文件大小
	if header.Size > constant.MaxUploadSize {
		http.Error(w, fmt.Sprintf("文件大小超过限制，最大允许%dMB", constant.MaxUploadSize/(1024*1024)), http.StatusBadRequest)
		return
	}

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	if len(fileData) == 0 {
		http.Error(w, "文件内容为空", http.StatusBadRequest)
		return
	}

	// 调用biz层上传固件
	filePath, err := s.uc.UploadFirmware(ctx, fileData, originalFilename)
	if err != nil {
		response := &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回成功响应（与Java实现一致，直接返回文件路径字符串）
	// Java的Result<String>序列化为JSON时，data字段直接是字符串值
	// 前端期望 res.data 直接是文件路径字符串
	// 由于protobuf Struct的限制，我们需要手动构建JSON响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 手动构建JSON响应，使data字段直接是字符串，与Java的Result<String>保持一致
	responseJSON := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": filePath, // 直接返回文件路径字符串
	}
	if err := json.NewEncoder(w).Encode(responseJSON); err != nil {
		// 编码失败，但响应头已发送，无法返回错误
		return
	}
}

// CheckOTAVersionHandler 处理 POST /ota/ - OTA版本和设备激活检查
func (s *OtaService) CheckOTAVersionHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// 从Header获取Device-Id和Client-Id
	deviceId := r.Header.Get("Device-Id")
	clientId := r.Header.Get("Client-Id")

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := &biz.DeviceReportRespDTO{
			Error: "Failed to read request body",
		}
		s.writeJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	// 解析JSON请求体
	var deviceReport biz.DeviceReportReqDTO
	if err := json.Unmarshal(body, &deviceReport); err != nil {
		response := &biz.DeviceReportRespDTO{
			Error: "Invalid request body",
		}
		s.writeJSONResponse(w, response, http.StatusBadRequest)
		return
	}

	// 调用service层内部方法
	response, err := s.checkOTAVersionInternal(ctx, deviceId, clientId, &deviceReport)
	if err != nil {
		response = &biz.DeviceReportRespDTO{
			Error: err.Error(),
		}
		s.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	// 返回JSON响应（使用JsonInclude.Include.NON_NULL策略，即omitempty）
	s.writeJSONResponse(w, response, http.StatusOK)
}

// ActivateDeviceHandler 处理 POST /ota/activate - 快速检查激活状态
func (s *OtaService) ActivateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// 从Header获取Device-Id
	deviceId := r.Header.Get("Device-Id")

	// 调用service层内部方法
	activated, err := s.activateDeviceInternal(ctx, deviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !activated {
		// 设备不存在，返回202状态码
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// 设备存在，返回"success"
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// GetOTAHealthHandler 处理 GET /ota - OTA健康检查
func (s *OtaService) GetOTAHealthHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// 调用service层内部方法
	message, err := s.getOTAHealthInternal(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("OTA接口异常"))
		return
	}

	// 返回文本响应
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

// writeJSONResponse 写入JSON响应（使用omitempty策略）
func (s *OtaService) writeJSONResponse(w http.ResponseWriter, response *biz.DeviceReportRespDTO, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	// 使用json.Marshal，Go的omitempty标签会自动处理空值
	jsonData, err := json.Marshal(response)
	if err != nil {
		// 如果序列化失败，返回错误响应
		errorResponse := map[string]string{
			"error": "Failed to serialize response",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Write(jsonData)
}

// RegisterOtaHTTPHandlers 注册OTA的HTTP handlers
func RegisterOtaHTTPHandlers(srv *kratoshttp.Server, otaService *OtaService) {
	// 注意：静态路由必须在动态路由之前注册，防止路由被覆盖
	// 1. 静态路由：/ota/activate (POST/OPTIONS)
	srv.HandleFunc("/ota/activate", func(w http.ResponseWriter, r *http.Request) {
		if handleCORS(w, r, "POST, OPTIONS", "Content-Type, Device-Id, Client-Id") {
			return
		}
		otaService.ActivateDeviceHandler(w, r)
	})

	// 2. 静态路由：/ota (GET/POST/OPTIONS)
	srv.HandleFunc("/ota", func(w http.ResponseWriter, r *http.Request) {
		if handleCORS(w, r, "GET, POST, OPTIONS", "Content-Type, Device-Id, Client-Id") {
			return
		}
		if r.Method == "GET" {
			otaService.GetOTAHealthHandler(w, r)
		} else if r.Method == "POST" {
			otaService.CheckOTAVersionHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 3. OTA固件管理相关路由（静态路由放在前面）
	srv.HandlePrefix("/otaMag/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleCORS(w, r, "GET, OPTIONS", "Content-Type") {
			return
		}
		otaService.DownloadOtaHandler(w, r)
	}))

	srv.HandleFunc("/otaMag/upload", func(w http.ResponseWriter, r *http.Request) {
		if handleCORS(w, r, "POST, OPTIONS", "Content-Type, Authorization") {
			return
		}
		otaService.UploadFirmwareHandler(w, r)
	})
}

// extractTokenFromRequest 从HTTP请求中提取Token
// 参考middleware实现：从Authorization header中提取Bearer Token
func extractTokenFromRequest(r *http.Request) string {
	auth := r.Header.Get(middleware.AuthorizationHeader)
	if auth == "" {
		return ""
	}

	// 检查Bearer格式：Authorization: Bearer {token}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

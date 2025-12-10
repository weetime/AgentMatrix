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

	"github.com/weetime/agent-matrix/internal/constant"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// DownloadOtaHandler 处理固件下载的HTTP handler
func (s *OtaService) DownloadOtaHandler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

// RegisterOtaHTTPHandlers 注册OTA的HTTP handlers
func RegisterOtaHTTPHandlers(srv *kratoshttp.Server, otaService *OtaService) {
	// 注意：静态路由（/download/{uuid} 和 /upload）必须在动态路由（/{id}）之前注册
	// 使用HandlePrefix注册自定义handler，匹配/otaMag/download/开头的所有路径
	srv.HandlePrefix("/otaMag/download/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otaService.DownloadOtaHandler(w, r)
	}))

	srv.HandleFunc("/otaMag/upload", func(w http.ResponseWriter, r *http.Request) {
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

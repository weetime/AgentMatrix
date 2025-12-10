package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/weetime/agent-matrix/internal/constant"
	"github.com/weetime/agent-matrix/internal/kit"
	pb "github.com/weetime/agent-matrix/protos/v1"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// UploadVoiceHandler 处理文件上传的HTTP handler
func (s *VoiceCloneService) UploadVoiceHandler(w http.ResponseWriter, r *http.Request) {
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

	// 解析multipart/form-data
	err := r.ParseMultipartForm(constant.MaxUploadSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
		return
	}

	// 获取id参数
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "id参数不能为空", http.StatusBadRequest)
		return
	}

	// 获取文件
	file, header, err := r.FormFile("voiceFile")
	if err != nil {
		http.Error(w, fmt.Sprintf("获取文件失败: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "audio/") {
		http.Error(w, "文件类型必须是音频文件", http.StatusBadRequest)
		return
	}

	// 验证文件大小
	if header.Size > constant.MaxUploadSize {
		http.Error(w, fmt.Sprintf("文件大小超过限制，最大允许%dMB", constant.MaxUploadSize/(1024*1024)), http.StatusBadRequest)
		return
	}

	// 读取文件数据
	voiceData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	if len(voiceData) == 0 {
		http.Error(w, "文件内容为空", http.StatusBadRequest)
		return
	}

	// 从请求中获取context（kratos会自动注入）
	ctx := r.Context()

	// 检查权限
	if err := s.checkPermission(ctx, id); err != nil {
		// 返回JSON格式的错误响应
		response := &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 调用service层上传音频
	err = s.uc.UploadVoice(ctx, id, voiceData)
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

	// 返回成功响应
	response := &pb.Response{
		Code: 0,
		Msg:  "success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PlayVoiceHandler 处理音频播放的HTTP handler
func (s *VoiceCloneService) PlayVoiceHandler(w http.ResponseWriter, r *http.Request) {
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
	// 路径格式: /voiceClone/play/{uuid}
	// 使用HandleFunc时，路径已经匹配到/voiceClone/play/，需要提取后面的uuid
	path := r.URL.Path
	pathPrefix := "/voiceClone/play/"
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

	// 从Redis获取音频ID
	redisKey := kit.GetVoiceCloneAudioIdKey(uuidStr)
	id, err := s.redisClient.Get(ctx, redisKey)
	if err != nil || id == "" {
		http.Error(w, "音频ID不存在或已过期", http.StatusNotFound)
		return
	}

	// 删除Redis key（一次性使用）
	_ = s.redisClient.Delete(ctx, redisKey)

	// 获取音频数据
	voiceData, err := s.uc.GetVoiceData(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取音频数据失败: %v", err), http.StatusInternalServerError)
		return
	}
	if voiceData == nil || len(voiceData) == 0 {
		http.Error(w, "音频数据不存在", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Length", strconv.Itoa(len(voiceData)))
	w.Header().Set("Content-Disposition", "inline; filename=voice.wav")

	// 写入音频数据
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(voiceData)
	if err != nil {
		// 写入失败，但响应头已发送，无法返回错误
		return
	}
}

// RegisterVoiceCloneHTTPHandlers 注册VoiceClone的HTTP handlers
func RegisterVoiceCloneHTTPHandlers(srv *kratoshttp.Server, voiceCloneService *VoiceCloneService) {
	// 注意：静态路由（/play/{uuid}）必须在动态路由（/audio/{id}）之前注册
	// 使用HandlePrefix注册自定义handler，匹配/voiceClone/play/开头的所有路径
	srv.HandlePrefix("/voiceClone/play/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		voiceCloneService.PlayVoiceHandler(w, r)
	}))

	srv.HandleFunc("/voiceClone/upload", func(w http.ResponseWriter, r *http.Request) {
		voiceCloneService.UploadVoiceHandler(w, r)
	})
}

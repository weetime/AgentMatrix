package biz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// VoiceDTO 音色DTO（用于返回给前端）
type VoiceDTO struct {
	ID        string // 音色ID
	Name      string // 音色名称
	VoiceDemo string // 音频播放地址（可选）
}

// VoiceClone 声音克隆实体
type VoiceClone struct {
	ID          string    // 唯一标识
	Name        string    // 声音名称
	ModelID     string    // 模型id
	VoiceID     string    // 声音id
	UserID      int64     // 用户ID（关联用户表）
	Voice       []byte    // 声音（音频数据）
	TrainStatus int32     // 训练状态：0待训练 1训练中 2训练成功 3训练失败
	TrainError  string    // 训练错误原因
	Creator     int64     // 创建者ID
	CreateDate  time.Time // 创建时间
}

// VoiceCloneResponseDTO 声音克隆响应DTO（用于向前端展示，包含模型名称和用户名称）
type VoiceCloneResponseDTO struct {
	ID          string // 唯一标识
	Name        string // 声音名称
	ModelID     string // 模型id
	ModelName   string // 模型名称（关联查询）
	VoiceID     string // 声音id
	UserID      string // 用户ID（格式化为字符串）
	UserName    string // 用户名称（关联查询）
	TrainStatus int32  // 训练状态：0待训练 1训练中 2训练成功 3训练失败
	TrainError  string // 训练错误原因
	CreateDate  string // 创建时间（格式化）
	HasVoice    bool   // 是否有音频数据
}

// ListVoiceCloneParams 分页查询参数
type ListVoiceCloneParams struct {
	UserID int64   // 用户ID（必填，用于权限过滤）
	Name   *string // 可选，声音名称或voiceId（模糊查询）
}

// VoiceCloneRepo 音色克隆数据访问接口
type VoiceCloneRepo interface {
	// GetTrainSuccess 获取用户训练成功的音色列表
	// modelId: 模型ID
	// userId: 用户ID
	GetTrainSuccess(ctx context.Context, modelId string, userId int64) ([]*VoiceDTO, error)

	// PageVoiceClone 分页查询声音克隆列表
	PageVoiceClone(ctx context.Context, params *ListVoiceCloneParams, page *kit.PageRequest) ([]*VoiceClone, int, error)

	// GetByID 根据ID查询声音克隆记录
	GetByID(ctx context.Context, id string) (*VoiceClone, error)

	// UpdateVoice 更新音频数据
	UpdateVoice(ctx context.Context, id string, voiceData []byte) error

	// UpdateName 更新声音克隆名称
	UpdateName(ctx context.Context, id string, name string) error

	// UpdateTrainStatus 更新训练状态
	UpdateTrainStatus(ctx context.Context, id string, status int32, errorMsg string) error

	// UpdateVoiceID 更新voiceId
	UpdateVoiceID(ctx context.Context, id string, voiceID string) error
}

// VoiceCloneUsecase 音色克隆业务逻辑
type VoiceCloneUsecase struct {
	repo        VoiceCloneRepo
	modelRepo   ModelConfigRepo
	userRepo    UserRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewVoiceCloneUsecase 创建音色克隆用例
func NewVoiceCloneUsecase(
	repo VoiceCloneRepo,
	modelRepo ModelConfigRepo,
	userRepo UserRepo,
	logger log.Logger,
) *VoiceCloneUsecase {
	return &VoiceCloneUsecase{
		repo:        repo,
		modelRepo:   modelRepo,
		userRepo:    userRepo,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// GetTrainSuccess 获取用户训练成功的音色列表
func (uc *VoiceCloneUsecase) GetTrainSuccess(ctx context.Context, modelId string, userId int64) ([]*VoiceDTO, error) {
	if modelId == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("modelId不能为空"))
	}
	if userId <= 0 {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("userId不能为空"))
	}
	return uc.repo.GetTrainSuccess(ctx, modelId, userId)
}

// PageVoiceClone 分页查询声音克隆列表（带模型名称和用户名称）
func (uc *VoiceCloneUsecase) PageVoiceClone(ctx context.Context, params *ListVoiceCloneParams, page *kit.PageRequest) ([]*VoiceCloneResponseDTO, int, error) {
	if params == nil || params.UserID <= 0 {
		return nil, 0, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("userId不能为空"))
	}

	// 查询列表
	list, total, err := uc.repo.PageVoiceClone(ctx, params, page)
	if err != nil {
		return nil, 0, err
	}

	// 转换为DTO列表
	dtoList := make([]*VoiceCloneResponseDTO, 0, len(list))
	for _, entity := range list {
		dto := uc.convertToResponseDTO(ctx, entity)
		dtoList = append(dtoList, dto)
	}

	return dtoList, total, nil
}

// GetVoiceCloneByID 根据ID查询声音克隆信息（带关联信息）
func (uc *VoiceCloneUsecase) GetVoiceCloneByID(ctx context.Context, id string) (*VoiceCloneResponseDTO, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("声音克隆记录不存在"))
	}

	return uc.convertToResponseDTO(ctx, entity), nil
}

// UploadVoice 上传音频文件
func (uc *VoiceCloneUsecase) UploadVoice(ctx context.Context, id string, voiceData []byte) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if len(voiceData) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("音频数据不能为空"))
	}

	// 验证记录存在
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("声音克隆记录不存在"))
	}

	// 更新音频数据和训练状态
	if err := uc.repo.UpdateVoice(ctx, id, voiceData); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 更新训练状态为待训练（0）
	if err := uc.repo.UpdateTrainStatus(ctx, id, 0, ""); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// UpdateName 更新声音克隆名称
func (uc *VoiceCloneUsecase) UpdateName(ctx context.Context, id string, name string) error {
	if id == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}
	if name == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("名称不能为空"))
	}

	// 验证记录存在
	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("声音克隆记录不存在"))
	}

	// 更新名称
	return uc.repo.UpdateName(ctx, id, name)
}

// GetVoiceData 获取音频数据
func (uc *VoiceCloneUsecase) GetVoiceData(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("id不能为空"))
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("声音克隆记录不存在"))
	}

	return entity.Voice, nil
}

// CloneAudio 克隆音频，调用火山引擎进行语音复刻训练
func (uc *VoiceCloneUsecase) CloneAudio(ctx context.Context, cloneId string) error {
	if cloneId == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("cloneId不能为空"))
	}

	// 查询声音克隆记录
	entity, err := uc.repo.GetByID(ctx, cloneId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if entity == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("声音克隆记录不存在"))
	}

	// 验证音频数据已上传
	if entity.Voice == nil || len(entity.Voice) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("音频数据未上传"))
	}

	// 获取模型配置
	modelConfig, err := uc.modelRepo.GetModelConfigByID(ctx, entity.ModelID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("获取模型配置失败: %w", err))
	}
	if modelConfig == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("模型配置不存在"))
	}

	// 调用火山引擎API进行语音克隆训练
	return uc.huoshanClone(ctx, modelConfig, entity)
}

// convertToResponseDTO 将VoiceClone实体转换为VoiceCloneResponseDTO
func (uc *VoiceCloneUsecase) convertToResponseDTO(ctx context.Context, entity *VoiceClone) *VoiceCloneResponseDTO {
	dto := &VoiceCloneResponseDTO{
		ID:          entity.ID,
		Name:        entity.Name,
		ModelID:     entity.ModelID,
		VoiceID:     entity.VoiceID,
		UserID:      fmt.Sprintf("%d", entity.UserID),
		TrainStatus: entity.TrainStatus,
		TrainError:  entity.TrainError,
		HasVoice:    entity.Voice != nil && len(entity.Voice) > 0,
	}

	// 格式化创建时间
	if !entity.CreateDate.IsZero() {
		dto.CreateDate = entity.CreateDate.Format("2006-01-02 15:04:05")
	}

	// 设置模型名称
	if entity.ModelID != "" {
		if config, err := uc.modelRepo.GetModelConfigByID(ctx, entity.ModelID); err == nil && config != nil {
			dto.ModelName = config.ModelName
		}
	}

	// 设置用户名称
	if entity.UserID > 0 {
		if user, err := uc.userRepo.GetByUserId(ctx, entity.UserID); err == nil && user != nil {
			dto.UserName = user.Username
		}
	}

	return dto
}

// GetUserByID 获取用户信息（辅助方法，用于UserUsecase接口）
func (uc *VoiceCloneUsecase) GetUserByID(ctx context.Context, userId int64) (*User, error) {
	return uc.userRepo.GetByUserId(ctx, userId)
}

// huoshanClone 调用火山引擎进行语音复刻训练
func (uc *VoiceCloneUsecase) huoshanClone(ctx context.Context, modelConfig *ModelConfig, entity *VoiceClone) error {
	// 解析模型配置JSON
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(modelConfig.ConfigJSON), &config); err != nil {
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("解析模型配置失败: %w", err))
	}

	// 获取appid和access_token
	appid, _ := config["appid"].(string)
	accessToken, _ := config["access_token"].(string)

	if appid == "" || accessToken == "" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("火山引擎配置缺失：appid或access_token不能为空"))
	}

	// 检查类型
	configType, _ := config["type"].(string)
	if configType != "huoshan_double_stream" {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("不支持的模型类型: %s", configType))
	}

	// 检查Voice ID格式（火山引擎双声道需要S_前缀）
	if entity.VoiceID != "" && !strings.HasPrefix(entity.VoiceID, "S_") {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("火山引擎双声道voiceId必须以S_开头"))
	}

	// 将音频数据Base64编码
	audioBase64 := base64.StdEncoding.EncodeToString(entity.Voice)

	// 构建请求体
	reqBody := map[string]interface{}{
		"appid": appid,
		"audios": []map[string]string{
			{
				"audio_bytes":  audioBase64,
				"audio_format": "wav",
			},
		},
		"source":     2,
		"language":   0,
		"model_type": 1,
		"speaker_id": entity.VoiceID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("构建请求体失败: %w", err))
	}

	// 调用火山引擎API
	apiUrl := "https://openspeech.bytedance.com/api/v1/mega_tts/audio/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, strings.NewReader(string(reqBodyJSON)))
	if err != nil {
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("创建HTTP请求失败: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer;"+accessToken)
	req.Header.Set("Resource-Id", "seed-icl-1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		// 更新训练状态为失败
		_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, fmt.Sprintf("请求失败: %v", err))
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("调用火山引擎API失败: %w", err))
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, fmt.Sprintf("读取响应失败: %v", err))
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("读取响应失败: %w", err))
	}

	uc.log.Infof("火山引擎API响应状态码: %d, 响应体: %s", resp.StatusCode, string(respBody))

	// 解析响应
	var rsp map[string]interface{}
	if err := json.Unmarshal(respBody, &rsp); err != nil {
		_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, fmt.Sprintf("解析响应失败: %v", err))
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("解析响应失败: %w", err))
	}

	// 获取BaseResp对象
	baseResp, ok := rsp["BaseResp"].(map[string]interface{})
	if ok && baseResp != nil {
		statusCode, _ := baseResp["StatusCode"].(float64)
		statusMessage, _ := baseResp["StatusMessage"].(string)

		// 获取speaker_id
		speakerId, _ := rsp["speaker_id"].(string)

		// StatusCode == 0 表示成功
		if statusCode == 0 && speakerId != "" {
			// 更新voiceId
			if err := uc.repo.UpdateVoiceID(ctx, entity.ID, speakerId); err != nil {
				return uc.handleError.ErrInternal(ctx, fmt.Errorf("更新voiceId失败: %w", err))
			}
			// 更新训练状态为成功
			if err := uc.repo.UpdateTrainStatus(ctx, entity.ID, 2, ""); err != nil {
				return uc.handleError.ErrInternal(ctx, fmt.Errorf("更新训练状态失败: %w", err))
			}
			return nil
		} else {
			// 失败时使用StatusMessage作为错误信息
			errorMsg := statusMessage
			if errorMsg == "" {
				errorMsg = "训练失败"
			}
			_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, errorMsg)
			return uc.handleError.ErrInternal(ctx, fmt.Errorf("训练失败: %s", errorMsg))
		}
	} else {
		// 尝试从message字段获取错误信息
		errorMsg, _ := rsp["message"].(string)
		if errorMsg != "" {
			_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, errorMsg)
			return uc.handleError.ErrInternal(ctx, fmt.Errorf("训练失败: %s", errorMsg))
		}
		_ = uc.repo.UpdateTrainStatus(ctx, entity.ID, 3, "响应格式错误")
		return uc.handleError.ErrInternal(ctx, fmt.Errorf("响应格式错误"))
	}
}

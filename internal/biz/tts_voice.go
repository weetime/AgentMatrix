package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// TtsVoice 音色业务实体
type TtsVoice struct {
	ID             string
	TtsModelID     string
	Name           string
	TtsVoice       string
	Languages      string
	VoiceDemo      string
	Remark         string
	ReferenceAudio string
	ReferenceText  string
	Sort           int32
	Creator        int64
	CreateDate     time.Time
	Updater        int64
	UpdateDate     time.Time
}

// ListTtsVoiceParams 查询音色过滤条件
type ListTtsVoiceParams struct {
	TtsModelID string  // TTS模型ID（必填）
	Name       *string // 音色名称（模糊查询）
}

// TtsVoiceRepo 音色数据访问接口
type TtsVoiceRepo interface {
	ListTtsVoice(ctx context.Context, params *ListTtsVoiceParams, page *kit.PageRequest) ([]*TtsVoice, error)
	TotalTtsVoice(ctx context.Context, params *ListTtsVoiceParams) (int, error)
	GetTtsVoiceByID(ctx context.Context, id string) (*TtsVoice, error)
	CreateTtsVoice(ctx context.Context, voice *TtsVoice) (*TtsVoice, error)
	UpdateTtsVoice(ctx context.Context, voice *TtsVoice) error
	DeleteTtsVoice(ctx context.Context, ids []string) error
}

// TtsVoiceUsecase 音色业务逻辑
type TtsVoiceUsecase struct {
	repo        TtsVoiceRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewTtsVoiceUsecase 创建音色用例
func NewTtsVoiceUsecase(
	repo TtsVoiceRepo,
	logger log.Logger,
) *TtsVoiceUsecase {
	return &TtsVoiceUsecase{
		repo:        repo,
		handleError: cerrors.NewHandleError(logger),
		log:         kit.LogHelper(logger),
	}
}

// ListTtsVoice 分页查询音色
func (uc *TtsVoiceUsecase) ListTtsVoice(ctx context.Context, params *ListTtsVoiceParams, page *kit.PageRequest) ([]*TtsVoice, error) {
	if err := kit.Validate(params); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}
	if params.TtsModelID == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("ttsModelId不能为空"))
	}
	return uc.repo.ListTtsVoice(ctx, params, page)
}

// TotalTtsVoice 获取音色总数
func (uc *TtsVoiceUsecase) TotalTtsVoice(ctx context.Context, params *ListTtsVoiceParams) (int, error) {
	if err := kit.Validate(params); err != nil {
		return 0, uc.handleError.ErrInvalidInput(ctx, err)
	}
	if params.TtsModelID == "" {
		return 0, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("ttsModelId不能为空"))
	}
	return uc.repo.TotalTtsVoice(ctx, params)
}

// GetTtsVoiceByID 根据ID获取音色
func (uc *TtsVoiceUsecase) GetTtsVoiceByID(ctx context.Context, id string) (*TtsVoice, error) {
	if id == "" {
		return nil, uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("音色ID不能为空"))
	}
	voice, err := uc.repo.GetTtsVoiceByID(ctx, id)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if voice == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("音色不存在"))
	}
	return voice, nil
}

// CreateTtsVoice 创建音色
func (uc *TtsVoiceUsecase) CreateTtsVoice(ctx context.Context, voice *TtsVoice) (*TtsVoice, error) {
	if err := kit.Validate(voice); err != nil {
		return nil, uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 设置默认值
	if voice.CreateDate.IsZero() {
		voice.CreateDate = time.Now()
	}
	if voice.UpdateDate.IsZero() {
		voice.UpdateDate = time.Now()
	}

	return uc.repo.CreateTtsVoice(ctx, voice)
}

// UpdateTtsVoice 更新音色
func (uc *TtsVoiceUsecase) UpdateTtsVoice(ctx context.Context, voice *TtsVoice) error {
	if err := kit.Validate(voice); err != nil {
		return uc.handleError.ErrInvalidInput(ctx, err)
	}

	// 检查音色是否存在
	existing, err := uc.repo.GetTtsVoiceByID(ctx, voice.ID)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if existing == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("音色不存在"))
	}

	// 设置更新时间
	voice.UpdateDate = time.Now()

	return uc.repo.UpdateTtsVoice(ctx, voice)
}

// DeleteTtsVoice 批量删除音色
func (uc *TtsVoiceUsecase) DeleteTtsVoice(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return uc.handleError.ErrInvalidInput(ctx, fmt.Errorf("音色ID列表不能为空"))
	}

	return uc.repo.DeleteTtsVoice(ctx, ids)
}

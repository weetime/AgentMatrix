package biz

import (
	"context"
	"fmt"

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

// VoiceCloneRepo 音色克隆数据访问接口
type VoiceCloneRepo interface {
	// GetTrainSuccess 获取用户训练成功的音色列表
	// modelId: 模型ID
	// userId: 用户ID
	GetTrainSuccess(ctx context.Context, modelId string, userId int64) ([]*VoiceDTO, error)
}

// VoiceCloneUsecase 音色克隆业务逻辑
type VoiceCloneUsecase struct {
	repo        VoiceCloneRepo
	handleError *cerrors.HandleError
	log         *log.Helper
}

// NewVoiceCloneUsecase 创建音色克隆用例
func NewVoiceCloneUsecase(
	repo VoiceCloneRepo,
	logger log.Logger,
) *VoiceCloneUsecase {
	return &VoiceCloneUsecase{
		repo:        repo,
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

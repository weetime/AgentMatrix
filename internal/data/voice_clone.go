package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent/voiceclone"

	"github.com/go-kratos/kratos/v2/log"
)

type voiceCloneRepo struct {
	data *Data
	log  *log.Helper
}

// NewVoiceCloneRepo 初始化音色克隆Repo
func NewVoiceCloneRepo(data *Data, logger log.Logger) biz.VoiceCloneRepo {
	return &voiceCloneRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/voice-clone")),
	}
}

// GetTrainSuccess 获取用户训练成功的音色列表
// 查询条件：model_id = modelId AND user_id = userId AND train_status = 2
// 返回字段：id, name, voice_id (作为 voiceDemo)
func (r *voiceCloneRepo) GetTrainSuccess(ctx context.Context, modelId string, userId int64) ([]*biz.VoiceDTO, error) {
	// 查询训练成功的克隆音色
	list, err := r.data.db.VoiceClone.Query().
		Where(
			voiceclone.ModelIDEQ(modelId),
			voiceclone.UserIDEQ(userId),
			voiceclone.TrainStatusEQ(2), // 2表示训练成功
		).
		Select(voiceclone.FieldID, voiceclone.FieldName, voiceclone.FieldVoiceID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.VoiceDTO, len(list))
	for i, item := range list {
		result[i] = &biz.VoiceDTO{
			ID:        item.ID,
			Name:      item.Name,
			VoiceDemo: item.VoiceID, // voice_id 作为 voiceDemo
		}
	}

	return result, nil
}

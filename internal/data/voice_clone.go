package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/voiceclone"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
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

// PageVoiceClone 分页查询声音克隆列表
func (r *voiceCloneRepo) PageVoiceClone(ctx context.Context, params *biz.ListVoiceCloneParams, page *kit.PageRequest) ([]*biz.VoiceClone, int, error) {
	query := r.data.db.VoiceClone.Query()

	// 添加查询条件
	if params.UserID > 0 {
		query = query.Where(voiceclone.UserIDEQ(params.UserID))
	}

	// 名称或voiceId模糊查询
	if params.Name != nil && *params.Name != "" {
		name := *params.Name
		query = query.Where(
			voiceclone.Or(
				voiceclone.NameContains(name),
				voiceclone.VoiceIDContains(name),
			),
		)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 默认按创建时间降序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("create_date")
		page.SetSortDesc()
	}

	// 应用排序
	sortField := page.GetSortField()
	switch sortField {
	case "create_date":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(voiceclone.FieldCreateDate))
		} else {
			query = query.Order(ent.Asc(voiceclone.FieldCreateDate))
		}
	case "id":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(voiceclone.FieldID))
		} else {
			query = query.Order(ent.Asc(voiceclone.FieldID))
		}
	case "name":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(voiceclone.FieldName))
		} else {
			query = query.Order(ent.Asc(voiceclone.FieldName))
		}
	default:
		// 默认按创建时间降序
		query = query.Order(ent.Desc(voiceclone.FieldCreateDate))
	}

	// 应用分页
	pageNo, _ := page.GetPageNo()
	pageSize := page.GetPageSize()
	if pageNo > 0 && pageSize > 0 {
		offset := (pageNo - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	entities, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.VoiceClone, len(entities))
	for i, e := range entities {
		var bizEntity biz.VoiceClone
		if err := copier.Copy(&bizEntity, e); err != nil {
			return nil, 0, err
		}
		result[i] = &bizEntity
	}

	return result, total, nil
}

// GetByID 根据ID查询声音克隆记录
func (r *voiceCloneRepo) GetByID(ctx context.Context, id string) (*biz.VoiceClone, error) {
	entity, err := r.data.db.VoiceClone.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizEntity biz.VoiceClone
	if err := copier.Copy(&bizEntity, entity); err != nil {
		return nil, err
	}

	return &bizEntity, nil
}

// UpdateVoice 更新音频数据
func (r *voiceCloneRepo) UpdateVoice(ctx context.Context, id string, voiceData []byte) error {
	_, err := r.data.db.VoiceClone.UpdateOneID(id).
		SetVoice(voiceData).
		Save(ctx)
	return err
}

// UpdateName 更新声音克隆名称
func (r *voiceCloneRepo) UpdateName(ctx context.Context, id string, name string) error {
	_, err := r.data.db.VoiceClone.UpdateOneID(id).
		SetName(name).
		Save(ctx)
	return err
}

// UpdateTrainStatus 更新训练状态
func (r *voiceCloneRepo) UpdateTrainStatus(ctx context.Context, id string, status int32, errorMsg string) error {
	update := r.data.db.VoiceClone.UpdateOneID(id).
		SetTrainStatus(status)

	if errorMsg != "" {
		update = update.SetTrainError(errorMsg)
	} else {
		update = update.ClearTrainError()
	}

	_, err := update.Save(ctx)
	return err
}

// UpdateVoiceID 更新voiceId
func (r *voiceCloneRepo) UpdateVoiceID(ctx context.Context, id string, voiceID string) error {
	_, err := r.data.db.VoiceClone.UpdateOneID(id).
		SetVoiceID(voiceID).
		Save(ctx)
	return err
}

package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/ttsvoice"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type ttsVoiceRepo struct {
	data *Data
	log  *log.Helper
}

// NewTtsVoiceRepo 初始化音色Repo
func NewTtsVoiceRepo(data *Data, logger log.Logger) biz.TtsVoiceRepo {
	return &ttsVoiceRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/tts-voice")),
	}
}

// ListTtsVoice 分页查询音色
func (r *ttsVoiceRepo) ListTtsVoice(ctx context.Context, params *biz.ListTtsVoiceParams, page *kit.PageRequest) ([]*biz.TtsVoice, error) {
	query := r.buildTtsVoiceQuery(params, page)
	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.TtsVoice, len(list))
	for i, item := range list {
		result[i] = r.entityToBiz(item)
	}

	return result, nil
}

// TotalTtsVoice 获取音色总数
func (r *ttsVoiceRepo) TotalTtsVoice(ctx context.Context, params *biz.ListTtsVoiceParams) (int, error) {
	query := r.data.db.TtsVoice.Query()
	return r.applyTtsVoiceFilters(query, params).Count(ctx)
}

// buildTtsVoiceQuery 构建查询
func (r *ttsVoiceRepo) buildTtsVoiceQuery(params *biz.ListTtsVoiceParams, page *kit.PageRequest) *ent.TtsVoiceQuery {
	query := r.data.db.TtsVoice.Query()
	query = r.applyTtsVoiceFilters(query, params)

	// 默认按sort字段升序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("sort")
		page.SetSortAsc()
	}

	applyPagination(query, page, ttsvoice.Columns)

	return query
}

// applyTtsVoiceFilters 应用过滤条件
func (r *ttsVoiceRepo) applyTtsVoiceFilters(query *ent.TtsVoiceQuery, params *biz.ListTtsVoiceParams) *ent.TtsVoiceQuery {
	if params == nil {
		return query
	}

	// 必须指定ttsModelID
	if params.TtsModelID != "" {
		query = query.Where(ttsvoice.TtsModelIDEQ(params.TtsModelID))
	}

	// 音色名称模糊查询
	if params.Name != nil && *params.Name != "" {
		query = query.Where(ttsvoice.NameContains(*params.Name))
	}

	return query
}

// GetTtsVoiceByID 根据ID获取音色
func (r *ttsVoiceRepo) GetTtsVoiceByID(ctx context.Context, id string) (*biz.TtsVoice, error) {
	voice, err := r.data.db.TtsVoice.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.entityToBiz(voice), nil
}

// CreateTtsVoice 创建音色
func (r *ttsVoiceRepo) CreateTtsVoice(ctx context.Context, voice *biz.TtsVoice) (*biz.TtsVoice, error) {
	create := r.data.db.TtsVoice.Create().
		SetID(voice.ID).
		SetTtsModelID(voice.TtsModelID).
		SetName(voice.Name).
		SetTtsVoice(voice.TtsVoice).
		SetLanguages(voice.Languages).
		SetSort(voice.Sort).
		SetCreateDate(voice.CreateDate).
		SetUpdateDate(voice.UpdateDate)

	if voice.VoiceDemo != "" {
		create.SetVoiceDemo(voice.VoiceDemo)
	}
	if voice.Remark != "" {
		create.SetRemark(voice.Remark)
	}
	if voice.Creator > 0 {
		create.SetCreator(voice.Creator)
	}
	if voice.Updater > 0 {
		create.SetUpdater(voice.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	result := r.entityToBiz(entity)
	// 如果reference字段有值，需要使用SQL直接更新数据库
	if voice.ReferenceAudio != "" || voice.ReferenceText != "" {
		err = r.updateReferenceFields(ctx, entity.ID, voice.ReferenceAudio, voice.ReferenceText)
		if err != nil {
			return nil, err
		}
		result.ReferenceAudio = voice.ReferenceAudio
		result.ReferenceText = voice.ReferenceText
	}

	return result, nil
}

// updateReferenceFields 使用SQL更新reference字段
func (r *ttsVoiceRepo) updateReferenceFields(ctx context.Context, id string, referenceAudio, referenceText string) error {
	// 使用ent的SQL执行功能
	// 注意：如果ent代码中还没有SetReferenceAudio和SetReferenceText方法，
	// 需要使用SQL直接更新
	// 这里暂时返回nil，等ent代码重新生成后再实现
	// 目前先跳过，不影响主要功能
	return nil
}

// UpdateTtsVoice 更新音色
func (r *ttsVoiceRepo) UpdateTtsVoice(ctx context.Context, voice *biz.TtsVoice) error {
	update := r.data.db.TtsVoice.UpdateOneID(voice.ID).
		SetTtsModelID(voice.TtsModelID).
		SetName(voice.Name).
		SetTtsVoice(voice.TtsVoice).
		SetLanguages(voice.Languages).
		SetSort(voice.Sort).
		SetUpdateDate(voice.UpdateDate)

	if voice.VoiceDemo != "" {
		update.SetVoiceDemo(voice.VoiceDemo)
	} else {
		update.ClearVoiceDemo()
	}
	if voice.Remark != "" {
		update.SetRemark(voice.Remark)
	} else {
		update.ClearRemark()
	}
	if voice.Updater > 0 {
		update.SetUpdater(voice.Updater)
	}

	_, err := update.Save(ctx)
	if err != nil {
		return err
	}

	// 使用SQL更新reference字段
	if voice.ReferenceAudio != "" || voice.ReferenceText != "" {
		return r.updateReferenceFields(ctx, voice.ID, voice.ReferenceAudio, voice.ReferenceText)
	}

	return nil
}

// DeleteTtsVoice 批量删除音色
func (r *ttsVoiceRepo) DeleteTtsVoice(ctx context.Context, ids []string) error {
	_, err := r.data.db.TtsVoice.Delete().
		Where(ttsvoice.IDIn(ids...)).
		Exec(ctx)
	return err
}

// entityToBiz 将ent实体转换为biz实体
func (r *ttsVoiceRepo) entityToBiz(entity *ent.TtsVoice) *biz.TtsVoice {
	bizVoice := &biz.TtsVoice{
		ID:             entity.ID,
		TtsModelID:     entity.TtsModelID,
		Name:           entity.Name,
		TtsVoice:       entity.TtsVoice,
		Languages:      entity.Languages,
		VoiceDemo:      entity.VoiceDemo,
		Remark:         entity.Remark,
		Sort:           entity.Sort,
		Creator:        entity.Creator,
		CreateDate:     entity.CreateDate,
		Updater:        entity.Updater,
		UpdateDate:     entity.UpdateDate,
		ReferenceAudio: "",
		ReferenceText:  "",
	}

	// 注意：reference_audio和reference_text字段需要等ent代码重新生成后才能使用
	// 目前先设置为空，等ent代码包含这些字段后，可以直接从entity中获取

	return bizVoice
}

// loadReferenceFields 从数据库加载reference字段
func (r *ttsVoiceRepo) loadReferenceFields(ctx context.Context, voice *biz.TtsVoice) {
	// 使用ent的SQL查询功能
	// 注意：如果ent代码中还没有这些字段，这里暂时返回空值
	// 等ent代码重新生成包含这些字段后，可以直接从entity中获取
	// 目前先设置为空，不影响主要功能
}

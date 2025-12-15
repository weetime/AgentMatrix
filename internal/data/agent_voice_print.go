package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/agentvoiceprint"

	"github.com/go-kratos/kratos/v2/log"
)

type agentVoicePrintRepo struct {
	data *Data
	log  *log.Helper
}

// NewAgentVoicePrintRepo 初始化 AgentVoicePrint Repo
func NewAgentVoicePrintRepo(data *Data, logger log.Logger) biz.AgentVoicePrintRepo {
	return &agentVoicePrintRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/agent_voice_print")),
	}
}

// CreateAgentVoicePrint 创建智能体声纹
func (r *agentVoicePrintRepo) CreateAgentVoicePrint(ctx context.Context, voicePrint *biz.AgentVoicePrint) (*biz.AgentVoicePrint, error) {
	create := r.data.db.AgentVoicePrint.Create().
		SetID(voicePrint.ID).
		SetAgentID(voicePrint.AgentID).
		SetAudioID(voicePrint.AudioID).
		SetSourceName(voicePrint.SourceName).
		SetCreateDate(voicePrint.CreateDate).
		SetUpdateDate(voicePrint.UpdateDate)

	if voicePrint.Introduce != "" {
		create = create.SetIntroduce(voicePrint.Introduce)
	}
	if voicePrint.Creator != 0 {
		create = create.SetCreator(voicePrint.Creator)
	}
	if voicePrint.Updater != 0 {
		create = create.SetUpdater(voicePrint.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.toBizAgentVoicePrint(entity), nil
}

// UpdateAgentVoicePrint 更新智能体声纹
func (r *agentVoicePrintRepo) UpdateAgentVoicePrint(ctx context.Context, voicePrint *biz.AgentVoicePrint) error {
	update := r.data.db.AgentVoicePrint.UpdateOneID(voicePrint.ID)

	if voicePrint.AudioID != "" {
		update = update.SetAudioID(voicePrint.AudioID)
	}
	if voicePrint.SourceName != "" {
		update = update.SetSourceName(voicePrint.SourceName)
	}
	if voicePrint.Introduce != "" {
		update = update.SetIntroduce(voicePrint.Introduce)
	}
	update = update.SetUpdateDate(voicePrint.UpdateDate)
	if voicePrint.Updater != 0 {
		update = update.SetUpdater(voicePrint.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteAgentVoicePrint 删除智能体声纹
func (r *agentVoicePrintRepo) DeleteAgentVoicePrint(ctx context.Context, id string, userId int64) error {
	_, err := r.data.db.AgentVoicePrint.Delete().
		Where(
			agentvoiceprint.IDEQ(id),
			agentvoiceprint.CreatorEQ(userId),
		).
		Exec(ctx)
	return err
}

// ListAgentVoicePrints 获取智能体声纹列表
func (r *agentVoicePrintRepo) ListAgentVoicePrints(ctx context.Context, agentId string, userId int64) ([]*biz.AgentVoicePrint, error) {
	entities, err := r.data.db.AgentVoicePrint.Query().
		Where(
			agentvoiceprint.AgentIDEQ(agentId),
			agentvoiceprint.CreatorEQ(userId),
		).
		Order(ent.Asc(agentvoiceprint.FieldCreateDate)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentVoicePrint, len(entities))
	for i, e := range entities {
		result[i] = r.toBizAgentVoicePrint(e)
	}

	return result, nil
}

// GetAgentVoicePrintByID 根据ID获取智能体声纹
func (r *agentVoicePrintRepo) GetAgentVoicePrintByID(ctx context.Context, id string) (*biz.AgentVoicePrint, error) {
	entity, err := r.data.db.AgentVoicePrint.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.toBizAgentVoicePrint(entity), nil
}

// ListAgentVoicePrintIDsByAgentID 获取智能体的所有声纹ID
func (r *agentVoicePrintRepo) ListAgentVoicePrintIDsByAgentID(ctx context.Context, agentId string) ([]string, error) {
	entities, err := r.data.db.AgentVoicePrint.Query().
		Where(agentvoiceprint.AgentIDEQ(agentId)).
		Select(agentvoiceprint.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(entities))
	for i, e := range entities {
		ids[i] = e.ID
	}

	return ids, nil
}

// ListAgentVoicePrintsByAgentID 根据智能体ID获取声纹列表（不需要用户权限验证）
func (r *agentVoicePrintRepo) ListAgentVoicePrintsByAgentID(ctx context.Context, agentId string) ([]*biz.AgentVoicePrint, error) {
	entities, err := r.data.db.AgentVoicePrint.Query().
		Where(agentvoiceprint.AgentIDEQ(agentId)).
		Order(ent.Asc(agentvoiceprint.FieldCreateDate)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.AgentVoicePrint, len(entities))
	for i, e := range entities {
		result[i] = r.toBizAgentVoicePrint(e)
	}

	return result, nil
}

// toBizAgentVoicePrint 转换为业务实体
func (r *agentVoicePrintRepo) toBizAgentVoicePrint(entity *ent.AgentVoicePrint) *biz.AgentVoicePrint {
	return &biz.AgentVoicePrint{
		ID:         entity.ID,
		AgentID:    entity.AgentID,
		AudioID:    entity.AudioID,
		SourceName: entity.SourceName,
		Introduce:  entity.Introduce,
		CreateDate: entity.CreateDate,
		Creator:    entity.Creator,
		UpdateDate: entity.UpdateDate,
		Updater:    entity.Updater,
	}
}

package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/sysparams"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
)

type configRepo struct {
	data *Data
	log  *log.Helper
}

// NewConfigRepo 初始化 Config Repo
func NewConfigRepo(data *Data, logger log.Logger) biz.ConfigRepo {
	return &configRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/config")),
	}
}

// ListAllParams 查询所有系统参数（param_type = 1）
func (r *configRepo) ListAllParams(ctx context.Context) ([]*biz.SysParam, error) {
	params, err := r.data.db.SysParams.Query().
		Where(sysparams.ParamTypeEQ(1)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.SysParam, len(params))
	for i, param := range params {
		result[i] = &biz.SysParam{
			ID:         param.ID,
			ParamCode:  param.ParamCode,
			ParamValue: param.ParamValue,
			ValueType:  param.ValueType,
			ParamType:  int8(param.ParamType),
			Remark:     param.Remark,
		}
	}

	return result, nil
}

// ListSysParams 列表查询参数（使用项目规范）
func (r *configRepo) ListSysParams(ctx context.Context, params *biz.ListSysParamsParams, page *kit.PageRequest) ([]*biz.SysParam, error) {
	list, err := r.buildSysParamsQuery(params, page).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.SysParam, len(list))
	if err := copier.Copy(&result, list); err != nil {
		return nil, err
	}

	return result, nil
}

// TotalSysParams 获取参数总数
func (r *configRepo) TotalSysParams(ctx context.Context, params *biz.ListSysParamsParams) (int, error) {
	query := r.data.db.SysParams.Query()
	return r.applySysParamsFilters(query, params).Count(ctx)
}

// buildSysParamsQuery 构建查询
func (r *configRepo) buildSysParamsQuery(params *biz.ListSysParamsParams, page *kit.PageRequest) *ent.SysParamsQuery {
	query := r.data.db.SysParams.Query()
	query = r.applySysParamsFilters(query, params)

	applyPagination(query, page, sysparams.Columns)

	return query
}

// applySysParamsFilters 应用过滤条件
func (r *configRepo) applySysParamsFilters(query *ent.SysParamsQuery, params *biz.ListSysParamsParams) *ent.SysParamsQuery {
	// 只查询非系统参数（param_type = 1）
	query = query.Where(sysparams.ParamTypeEQ(1))

	// 按参数编码或备注查询（支持模糊查询）
	if params.ParamCode != nil && params.ParamCode.GetValue() != "" {
		query = query.Where(
			sysparams.Or(
				sysparams.ParamCodeContains(params.ParamCode.GetValue()),
				sysparams.RemarkContains(params.ParamCode.GetValue()),
			),
		)
	}

	return query
}

// PageSysParams 分页查询参数（兼容旧接口）
func (r *configRepo) PageSysParams(ctx context.Context, page, limit int32, paramCode string) ([]*biz.SysParam, int, error) {
	query := r.data.db.SysParams.Query().
		Where(sysparams.ParamTypeEQ(1)) // 只查询非系统参数（param_type = 1）

	// 按参数编码或备注查询（支持模糊查询）
	if paramCode != "" {
		query = query.Where(
			sysparams.Or(
				sysparams.ParamCodeContains(paramCode),
				sysparams.RemarkContains(paramCode),
			),
		)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	params, err := query.
		Offset(int(offset)).
		Limit(int(limit)).
		Order(ent.Desc(sysparams.FieldID)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.SysParam, len(params))
	for i, param := range params {
		result[i] = &biz.SysParam{
			ID:         param.ID,
			ParamCode:  param.ParamCode,
			ParamValue: param.ParamValue,
			ValueType:  param.ValueType,
			ParamType:  int8(param.ParamType),
			Remark:     param.Remark,
		}
	}

	return result, total, nil
}

// GetSysParamsByID 根据ID获取参数
func (r *configRepo) GetSysParamsByID(ctx context.Context, id int64) (*biz.SysParam, error) {
	param, err := r.data.db.SysParams.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &biz.SysParam{
		ID:         param.ID,
		ParamCode:  param.ParamCode,
		ParamValue: param.ParamValue,
		ValueType:  param.ValueType,
		ParamType:  int8(param.ParamType),
		Remark:     param.Remark,
	}, nil
}

// CreateSysParams 创建参数
func (r *configRepo) CreateSysParams(ctx context.Context, param *biz.SysParam) (*biz.SysParam, error) {
	// 生成ID（使用雪花算法）
	id := kit.GenerateInt64ID()

	create := r.data.db.SysParams.Create().
		SetID(id).
		SetParamCode(param.ParamCode).
		SetParamValue(param.ParamValue).
		SetValueType(param.ValueType).
		SetParamType(int8(param.ParamType)).
		SetNillableRemark(&param.Remark)

	if param.Creator > 0 {
		create.SetCreator(param.Creator)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.SysParam{
		ID:         entity.ID,
		ParamCode:  entity.ParamCode,
		ParamValue: entity.ParamValue,
		ValueType:  entity.ValueType,
		ParamType:  int8(entity.ParamType),
		Remark:     entity.Remark,
	}, nil
}

// UpdateSysParams 更新参数
func (r *configRepo) UpdateSysParams(ctx context.Context, param *biz.SysParam) error {
	update := r.data.db.SysParams.UpdateOneID(param.ID).
		SetParamCode(param.ParamCode).
		SetParamValue(param.ParamValue).
		SetValueType(param.ValueType).
		SetNillableRemark(&param.Remark)

	if param.Updater > 0 {
		update.SetUpdater(param.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteSysParams 批量删除参数
func (r *configRepo) DeleteSysParams(ctx context.Context, ids []int64) error {
	_, err := r.data.db.SysParams.Delete().
		Where(sysparams.IDIn(ids...)).
		Exec(ctx)
	return err
}

// GetSysParamsByCode 根据编码获取参数
func (r *configRepo) GetSysParamsByCode(ctx context.Context, paramCode string) (*biz.SysParam, error) {
	param, err := r.data.db.SysParams.Query().
		Where(sysparams.ParamCodeEQ(paramCode)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.SysParam{
		ID:         param.ID,
		ParamCode:  param.ParamCode,
		ParamValue: param.ParamValue,
		ValueType:  param.ValueType,
		ParamType:  int8(param.ParamType),
		Remark:     param.Remark,
	}, nil
}

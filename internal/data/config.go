package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent/sysparams"

	"github.com/go-kratos/kratos/v2/log"
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

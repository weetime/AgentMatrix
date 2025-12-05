package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/sysdictdata"
	"github.com/weetime/agent-matrix/internal/data/ent/sysdicttype"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type dictTypeRepo struct {
	data *Data
	log  *log.Helper
}

type dictDataRepo struct {
	data *Data
	log  *log.Helper
}

// NewDictTypeRepo 初始化字典类型Repo
func NewDictTypeRepo(data *Data, logger log.Logger) biz.DictTypeRepo {
	return &dictTypeRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/dict-type")),
	}
}

// NewDictDataRepo 初始化字典数据Repo
func NewDictDataRepo(data *Data, logger log.Logger) biz.DictDataRepo {
	return &dictDataRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/dict-data")),
	}
}

// ListDictTypes 分页查询字典类型
func (r *dictTypeRepo) ListDictTypes(ctx context.Context, params *biz.ListDictTypeParams, page *kit.PageRequest) ([]*biz.SysDictType, error) {
	query := r.buildDictTypeQuery(params, page)
	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.SysDictType, len(list))
	for i, item := range list {
		result[i] = &biz.SysDictType{
			ID:         item.ID,
			DictType:   item.DictType,
			DictName:   item.DictName,
			Remark:     item.Remark,
			Sort:       item.Sort,
			Creator:    item.Creator,
			CreateDate: item.CreateDate,
			Updater:    item.Updater,
			UpdateDate: item.UpdateDate,
		}
	}

	return result, nil
}

// TotalDictTypes 获取字典类型总数
func (r *dictTypeRepo) TotalDictTypes(ctx context.Context, params *biz.ListDictTypeParams) (int, error) {
	query := r.data.db.SysDictType.Query()
	return r.applyDictTypeFilters(query, params).Count(ctx)
}

// buildDictTypeQuery 构建查询
func (r *dictTypeRepo) buildDictTypeQuery(params *biz.ListDictTypeParams, page *kit.PageRequest) *ent.SysDictTypeQuery {
	query := r.data.db.SysDictType.Query()
	query = r.applyDictTypeFilters(query, params)

	// 默认按sort字段升序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("sort")
		page.SetSortAsc()
	}

	applyPagination(query, page, sysdicttype.Columns)

	return query
}

// applyDictTypeFilters 应用过滤条件
func (r *dictTypeRepo) applyDictTypeFilters(query *ent.SysDictTypeQuery, params *biz.ListDictTypeParams) *ent.SysDictTypeQuery {
	if params == nil {
		return query
	}

	if params.DictType != nil && *params.DictType != "" {
		query = query.Where(sysdicttype.DictTypeContains(*params.DictType))
	}

	if params.DictName != nil && *params.DictName != "" {
		query = query.Where(sysdicttype.DictNameContains(*params.DictName))
	}

	return query
}

// GetDictTypeByID 根据ID获取字典类型
func (r *dictTypeRepo) GetDictTypeByID(ctx context.Context, id int64) (*biz.SysDictType, error) {
	dictType, err := r.data.db.SysDictType.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.SysDictType{
		ID:         dictType.ID,
		DictType:   dictType.DictType,
		DictName:   dictType.DictName,
		Remark:     dictType.Remark,
		Sort:       dictType.Sort,
		Creator:    dictType.Creator,
		CreateDate: dictType.CreateDate,
		Updater:    dictType.Updater,
		UpdateDate: dictType.UpdateDate,
	}, nil
}

// GetDictTypeByType 根据字典类型编码获取字典类型
func (r *dictTypeRepo) GetDictTypeByType(ctx context.Context, dictType string) (*biz.SysDictType, error) {
	dt, err := r.data.db.SysDictType.Query().
		Where(sysdicttype.DictTypeEQ(dictType)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.SysDictType{
		ID:         dt.ID,
		DictType:   dt.DictType,
		DictName:   dt.DictName,
		Remark:     dt.Remark,
		Sort:       dt.Sort,
		Creator:    dt.Creator,
		CreateDate: dt.CreateDate,
		Updater:    dt.Updater,
		UpdateDate: dt.UpdateDate,
	}, nil
}

// CreateDictType 创建字典类型
func (r *dictTypeRepo) CreateDictType(ctx context.Context, dictType *biz.SysDictType) (*biz.SysDictType, error) {
	id := kit.GenerateInt64ID()

	create := r.data.db.SysDictType.Create().
		SetID(id).
		SetDictType(dictType.DictType).
		SetDictName(dictType.DictName).
		SetSort(dictType.Sort).
		SetCreateDate(dictType.CreateDate).
		SetUpdateDate(dictType.UpdateDate)

	if dictType.Remark != "" {
		create.SetRemark(dictType.Remark)
	}
	if dictType.Creator > 0 {
		create.SetCreator(dictType.Creator)
	}
	if dictType.Updater > 0 {
		create.SetUpdater(dictType.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.SysDictType{
		ID:         entity.ID,
		DictType:   entity.DictType,
		DictName:   entity.DictName,
		Remark:     entity.Remark,
		Sort:       entity.Sort,
		Creator:    entity.Creator,
		CreateDate: entity.CreateDate,
		Updater:    entity.Updater,
		UpdateDate: entity.UpdateDate,
	}, nil
}

// UpdateDictType 更新字典类型
func (r *dictTypeRepo) UpdateDictType(ctx context.Context, dictType *biz.SysDictType) error {
	update := r.data.db.SysDictType.UpdateOneID(dictType.ID).
		SetDictType(dictType.DictType).
		SetDictName(dictType.DictName).
		SetSort(dictType.Sort).
		SetUpdateDate(dictType.UpdateDate)

	if dictType.Remark != "" {
		update.SetRemark(dictType.Remark)
	} else {
		update.ClearRemark()
	}
	if dictType.Updater > 0 {
		update.SetUpdater(dictType.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteDictTypes 批量删除字典类型
func (r *dictTypeRepo) DeleteDictTypes(ctx context.Context, ids []int64) error {
	_, err := r.data.db.SysDictType.Delete().
		Where(sysdicttype.IDIn(ids...)).
		Exec(ctx)
	return err
}

// ListDictData 分页查询字典数据
func (r *dictDataRepo) ListDictData(ctx context.Context, params *biz.ListDictDataParams, page *kit.PageRequest) ([]*biz.SysDictData, error) {
	query := r.buildDictDataQuery(params, page)
	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.SysDictData, len(list))
	for i, item := range list {
		result[i] = &biz.SysDictData{
			ID:         item.ID,
			DictTypeID: item.DictTypeID,
			DictLabel:  item.DictLabel,
			DictValue:  item.DictValue,
			Remark:     item.Remark,
			Sort:       item.Sort,
			Creator:    item.Creator,
			CreateDate: item.CreateDate,
			Updater:    item.Updater,
			UpdateDate: item.UpdateDate,
		}
	}

	return result, nil
}

// TotalDictData 获取字典数据总数
func (r *dictDataRepo) TotalDictData(ctx context.Context, params *biz.ListDictDataParams) (int, error) {
	query := r.data.db.SysDictData.Query()
	return r.applyDictDataFilters(query, params).Count(ctx)
}

// buildDictDataQuery 构建查询
func (r *dictDataRepo) buildDictDataQuery(params *biz.ListDictDataParams, page *kit.PageRequest) *ent.SysDictDataQuery {
	query := r.data.db.SysDictData.Query()
	query = r.applyDictDataFilters(query, params)

	// 默认按sort字段升序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("sort")
		page.SetSortAsc()
	}

	applyPagination(query, page, sysdictdata.Columns)

	return query
}

// applyDictDataFilters 应用过滤条件
func (r *dictDataRepo) applyDictDataFilters(query *ent.SysDictDataQuery, params *biz.ListDictDataParams) *ent.SysDictDataQuery {
	if params == nil {
		return query
	}

	// 必须指定dictTypeID
	if params.DictTypeID > 0 {
		query = query.Where(sysdictdata.DictTypeIDEQ(params.DictTypeID))
	}

	if params.DictLabel != nil && *params.DictLabel != "" {
		query = query.Where(sysdictdata.DictLabelContains(*params.DictLabel))
	}

	if params.DictValue != nil && *params.DictValue != "" {
		query = query.Where(sysdictdata.DictValueContains(*params.DictValue))
	}

	return query
}

// GetDictDataByID 根据ID获取字典数据
func (r *dictDataRepo) GetDictDataByID(ctx context.Context, id int64) (*biz.SysDictData, error) {
	dictData, err := r.data.db.SysDictData.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.SysDictData{
		ID:         dictData.ID,
		DictTypeID: dictData.DictTypeID,
		DictLabel:  dictData.DictLabel,
		DictValue:  dictData.DictValue,
		Remark:     dictData.Remark,
		Sort:       dictData.Sort,
		Creator:    dictData.Creator,
		CreateDate: dictData.CreateDate,
		Updater:    dictData.Updater,
		UpdateDate: dictData.UpdateDate,
	}, nil
}

// GetDictDataByType 根据字典类型编码获取数据列表
func (r *dictDataRepo) GetDictDataByType(ctx context.Context, dictType string) ([]*biz.SysDictDataItem, error) {
	// 先查询字典类型ID
	dictTypeEntity, err := r.data.db.SysDictType.Query().
		Where(sysdicttype.DictTypeEQ(dictType)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// 查询字典数据，按sort升序排序
	dictDataList, err := r.data.db.SysDictData.Query().
		Where(sysdictdata.DictTypeIDEQ(dictTypeEntity.ID)).
		Order(ent.Asc(sysdictdata.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.SysDictDataItem, len(dictDataList))
	for i, item := range dictDataList {
		result[i] = &biz.SysDictDataItem{
			Name: item.DictLabel,
			Key:  item.DictValue,
		}
	}

	return result, nil
}

// GetTypeByTypeID 根据字典类型ID获取字典类型编码
func (r *dictDataRepo) GetTypeByTypeID(ctx context.Context, dictTypeID int64) (string, error) {
	dictType, err := r.data.db.SysDictType.Get(ctx, dictTypeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return dictType.DictType, nil
}

// CreateDictData 创建字典数据
func (r *dictDataRepo) CreateDictData(ctx context.Context, dictData *biz.SysDictData) (*biz.SysDictData, error) {
	id := kit.GenerateInt64ID()

	create := r.data.db.SysDictData.Create().
		SetID(id).
		SetDictTypeID(dictData.DictTypeID).
		SetDictLabel(dictData.DictLabel).
		SetSort(dictData.Sort).
		SetCreateDate(dictData.CreateDate).
		SetUpdateDate(dictData.UpdateDate)

	if dictData.DictValue != "" {
		create.SetDictValue(dictData.DictValue)
	}
	if dictData.Remark != "" {
		create.SetRemark(dictData.Remark)
	}
	if dictData.Creator > 0 {
		create.SetCreator(dictData.Creator)
	}
	if dictData.Updater > 0 {
		create.SetUpdater(dictData.Updater)
	}

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &biz.SysDictData{
		ID:         entity.ID,
		DictTypeID: entity.DictTypeID,
		DictLabel:  entity.DictLabel,
		DictValue:  entity.DictValue,
		Remark:     entity.Remark,
		Sort:       entity.Sort,
		Creator:    entity.Creator,
		CreateDate: entity.CreateDate,
		Updater:    entity.Updater,
		UpdateDate: entity.UpdateDate,
	}, nil
}

// UpdateDictData 更新字典数据
func (r *dictDataRepo) UpdateDictData(ctx context.Context, dictData *biz.SysDictData) error {
	update := r.data.db.SysDictData.UpdateOneID(dictData.ID).
		SetDictTypeID(dictData.DictTypeID).
		SetDictLabel(dictData.DictLabel).
		SetSort(dictData.Sort).
		SetUpdateDate(dictData.UpdateDate)

	if dictData.DictValue != "" {
		update.SetDictValue(dictData.DictValue)
	} else {
		update.ClearDictValue()
	}
	if dictData.Remark != "" {
		update.SetRemark(dictData.Remark)
	} else {
		update.ClearRemark()
	}
	if dictData.Updater > 0 {
		update.SetUpdater(dictData.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// DeleteDictData 批量删除字典数据
func (r *dictDataRepo) DeleteDictData(ctx context.Context, ids []int64) error {
	_, err := r.data.db.SysDictData.Delete().
		Where(sysdictdata.IDIn(ids...)).
		Exec(ctx)
	return err
}

// DeleteDictDataByTypeID 根据字典类型ID删除字典数据
func (r *dictDataRepo) DeleteDictDataByTypeID(ctx context.Context, dictTypeID int64) error {
	_, err := r.data.db.SysDictData.Delete().
		Where(sysdictdata.DictTypeIDEQ(dictTypeID)).
		Exec(ctx)
	return err
}

// CheckDictValueUnique 检查字典值是否唯一
func (r *dictDataRepo) CheckDictValueUnique(ctx context.Context, dictTypeID int64, dictValue string, excludeID *int64) (bool, error) {
	query := r.data.db.SysDictData.Query().
		Where(
			sysdictdata.DictTypeIDEQ(dictTypeID),
			sysdictdata.DictValueEQ(dictValue),
		)

	if excludeID != nil {
		query = query.Where(sysdictdata.IDNEQ(*excludeID))
	}

	exists, err := query.Exist(ctx)
	return exists, err
}

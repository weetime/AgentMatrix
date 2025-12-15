package data

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/ota"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
)

type otaRepo struct {
	data *Data
	log  *log.Helper
}

// NewOtaRepo 初始化OTA固件Repo
func NewOtaRepo(data *Data, logger log.Logger) biz.OtaRepo {
	return &otaRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/ota")),
	}
}

// PageOta 分页查询OTA固件列表
func (r *otaRepo) PageOta(ctx context.Context, params *biz.ListOtaParams, page *kit.PageRequest) ([]*biz.Ota, int, error) {
	query := r.data.db.Ota.Query()

	// 添加查询条件：固件名称模糊查询
	if params.FirmwareName != nil && *params.FirmwareName != "" {
		name := *params.FirmwareName
		query = query.Where(ota.FirmwareNameContains(name))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 默认按更新时间降序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("update_date")
		page.SetSortDesc()
	}

	// 应用排序
	sortField := page.GetSortField()
	switch sortField {
	case "update_date":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ota.FieldUpdateDate))
		} else {
			query = query.Order(ent.Asc(ota.FieldUpdateDate))
		}
	case "create_date":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ota.FieldCreateDate))
		} else {
			query = query.Order(ent.Asc(ota.FieldCreateDate))
		}
	case "id":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ota.FieldID))
		} else {
			query = query.Order(ent.Asc(ota.FieldID))
		}
	default:
		// 默认按更新时间降序
		query = query.Order(ent.Desc(ota.FieldUpdateDate))
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

	result := make([]*biz.Ota, len(entities))
	for i, e := range entities {
		var bizEntity biz.Ota
		if err := copier.Copy(&bizEntity, e); err != nil {
			return nil, 0, err
		}
		result[i] = &bizEntity
	}

	return result, total, nil
}

// GetByID 根据ID查询OTA固件记录
func (r *otaRepo) GetByID(ctx context.Context, id string) (*biz.Ota, error) {
	entity, err := r.data.db.Ota.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizEntity biz.Ota
	if err := copier.Copy(&bizEntity, entity); err != nil {
		return nil, err
	}

	return &bizEntity, nil
}

// Save 保存OTA固件（同类固件只保留最新一条）
func (r *otaRepo) Save(ctx context.Context, entity *biz.Ota) error {
	// 查询同类固件（相同type）
	existingList, err := r.data.db.Ota.Query().
		Where(ota.TypeEQ(entity.Type)).
		All(ctx)
	if err != nil {
		return err
	}

	// 如果存在同类固件，更新第一条记录
	if len(existingList) > 0 {
		existing := existingList[0]
		update := r.data.db.Ota.UpdateOneID(existing.ID).
			SetFirmwareName(entity.FirmwareName).
			SetType(entity.Type).
			SetVersion(entity.Version).
			SetUpdateDate(time.Now())

		if entity.Size > 0 {
			update = update.SetSize(entity.Size)
		}
		if entity.Remark != "" {
			update = update.SetRemark(entity.Remark)
		}
		if entity.FirmwarePath != "" {
			update = update.SetFirmwarePath(entity.FirmwarePath)
		}
		if entity.Sort > 0 {
			update = update.SetSort(entity.Sort)
		}
		if entity.Updater > 0 {
			update = update.SetUpdater(entity.Updater)
		}

		_, err = update.Save(ctx)
		return err
	}

	// 如果不存在，创建新记录
	create := r.data.db.Ota.Create().
		SetFirmwareName(entity.FirmwareName).
		SetType(entity.Type).
		SetVersion(entity.Version)

	// 生成ID（UUID去掉横线，限制32字符）
	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	if len(id) > 32 {
		id = id[:32]
	}
	create = create.SetID(id)

	if entity.Size > 0 {
		create = create.SetSize(entity.Size)
	}
	if entity.Remark != "" {
		create = create.SetRemark(entity.Remark)
	}
	if entity.FirmwarePath != "" {
		create = create.SetFirmwarePath(entity.FirmwarePath)
	}
	if entity.Sort > 0 {
		create = create.SetSort(entity.Sort)
	}
	if entity.Creator > 0 {
		create = create.SetCreator(entity.Creator)
	}
	if entity.Updater > 0 {
		create = create.SetUpdater(entity.Updater)
	}
	if !entity.CreateDate.IsZero() {
		create = create.SetCreateDate(entity.CreateDate)
	} else {
		create = create.SetCreateDate(time.Now())
	}
	if !entity.UpdateDate.IsZero() {
		create = create.SetUpdateDate(entity.UpdateDate)
	} else {
		create = create.SetUpdateDate(time.Now())
	}

	_, err = create.Save(ctx)
	return err
}

// Update 更新OTA固件（检查类型和版本唯一性）
func (r *otaRepo) Update(ctx context.Context, entity *biz.Ota) error {
	// 检查是否存在相同类型和版本的固件（排除当前记录）
	// 先查询所有匹配的记录
	list, err := r.data.db.Ota.Query().
		Where(
			ota.TypeEQ(entity.Type),
			ota.VersionEQ(entity.Version),
		).
		All(ctx)
	if err != nil {
		return err
	}

	// 过滤掉当前记录
	for _, item := range list {
		if item.ID != entity.ID {
			return biz.ErrDuplicateOtaTypeVersion
		}
	}

	update := r.data.db.Ota.UpdateOneID(entity.ID).
		SetFirmwareName(entity.FirmwareName).
		SetType(entity.Type).
		SetVersion(entity.Version).
		SetUpdateDate(time.Now())

	if entity.Size > 0 {
		update = update.SetSize(entity.Size)
	}
	if entity.Remark != "" {
		update = update.SetRemark(entity.Remark)
	}
	if entity.FirmwarePath != "" {
		update = update.SetFirmwarePath(entity.FirmwarePath)
	}
	if entity.Sort > 0 {
		update = update.SetSort(entity.Sort)
	}
	if entity.Updater > 0 {
		update = update.SetUpdater(entity.Updater)
	}

	_, err = update.Save(ctx)
	return err
}

// Delete 批量删除OTA固件
func (r *otaRepo) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := r.data.db.Ota.Delete().
		Where(ota.IDIn(ids...)).
		Exec(ctx)
	return err
}

// GetLatestOta 根据类型获取最新OTA固件
func (r *otaRepo) GetLatestOta(ctx context.Context, otaType string) (*biz.Ota, error) {
	if otaType == "" {
		return nil, nil
	}

	entity, err := r.data.db.Ota.Query().
		Where(ota.TypeEQ(otaType)).
		Order(ent.Desc(ota.FieldUpdateDate)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizEntity biz.Ota
	if err := copier.Copy(&bizEntity, entity); err != nil {
		return nil, err
	}

	return &bizEntity, nil
}

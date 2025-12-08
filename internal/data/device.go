package data

import (
	"context"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/device"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type deviceRepo struct {
	data *Data
	log  *log.Helper
}

// NewDeviceRepo 初始化设备Repo
func NewDeviceRepo(data *Data, logger log.Logger) biz.DeviceRepo {
	return &deviceRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/device")),
	}
}

// Create 创建设备
func (r *deviceRepo) Create(ctx context.Context, d *biz.Device) error {
	create := r.data.db.Device.Create().
		SetID(d.ID).
		SetUserID(d.UserID).
		SetCreateDate(d.CreateDate).
		SetUpdateDate(d.UpdateDate).
		SetAutoUpdate(d.AutoUpdate)

	if d.MacAddress != "" {
		create.SetMACAddress(d.MacAddress)
	}
	if d.LastConnectedAt != nil {
		create.SetLastConnectedAt(*d.LastConnectedAt)
	}
	if d.Board != "" {
		create.SetBoard(d.Board)
	}
	if d.Alias != "" {
		create.SetAlias(d.Alias)
	}
	if d.AgentID != "" {
		create.SetAgentID(d.AgentID)
	}
	if d.AppVersion != "" {
		create.SetAppVersion(d.AppVersion)
	}
	if d.Sort > 0 {
		create.SetSort(d.Sort)
	}
	if d.Creator > 0 {
		create.SetCreator(d.Creator)
	}
	if d.Updater > 0 {
		create.SetUpdater(d.Updater)
	}

	_, err := create.Save(ctx)
	return err
}

// Update 更新设备
func (r *deviceRepo) Update(ctx context.Context, d *biz.Device) error {
	update := r.data.db.Device.UpdateOneID(d.ID).
		SetUpdateDate(d.UpdateDate).
		SetAutoUpdate(d.AutoUpdate)

	if d.MacAddress != "" {
		update.SetMACAddress(d.MacAddress)
	}
	if d.LastConnectedAt != nil {
		update.SetLastConnectedAt(*d.LastConnectedAt)
	}
	if d.Board != "" {
		update.SetBoard(d.Board)
	}
	if d.Alias != "" {
		update.SetAlias(d.Alias)
	}
	if d.AgentID != "" {
		update.SetAgentID(d.AgentID)
	}
	if d.AppVersion != "" {
		update.SetAppVersion(d.AppVersion)
	}
	if d.Sort > 0 {
		update.SetSort(d.Sort)
	}
	if d.Updater > 0 {
		update.SetUpdater(d.Updater)
	}

	_, err := update.Save(ctx)
	return err
}

// Delete 删除设备
func (r *deviceRepo) Delete(ctx context.Context, deviceId string, userId int64) error {
	_, err := r.data.db.Device.Delete().
		Where(
			device.IDEQ(deviceId),
			device.UserIDEQ(userId),
		).
		Exec(ctx)
	return err
}

// GetByID 根据ID获取设备
func (r *deviceRepo) GetByID(ctx context.Context, deviceId string) (*biz.Device, error) {
	entity, err := r.data.db.Device.Get(ctx, deviceId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.toBizDevice(entity), nil
}

// GetByMacAddress 根据MAC地址获取设备
func (r *deviceRepo) GetByMacAddress(ctx context.Context, macAddress string) (*biz.Device, error) {
	entity, err := r.data.db.Device.Query().
		Where(device.MACAddressEQ(macAddress)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.toBizDevice(entity), nil
}

// ListByUserAndAgent 根据用户ID和智能体ID获取设备列表
func (r *deviceRepo) ListByUserAndAgent(ctx context.Context, userId int64, agentId string) ([]*biz.Device, error) {
	entities, err := r.data.db.Device.Query().
		Where(
			device.UserIDEQ(userId),
			device.AgentIDEQ(agentId),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.Device, len(entities))
	for i, entity := range entities {
		result[i] = r.toBizDevice(entity)
	}

	return result, nil
}

// toBizDevice 将Ent实体转换为Biz实体
func (r *deviceRepo) toBizDevice(entity *ent.Device) *biz.Device {
	d := &biz.Device{
		ID:         entity.ID,
		UserID:     entity.UserID,
		MacAddress: entity.MACAddress,
		AutoUpdate: entity.AutoUpdate,
		Board:      entity.Board,
		Alias:      entity.Alias,
		AgentID:    entity.AgentID,
		AppVersion: entity.AppVersion,
		Sort:       entity.Sort,
		Creator:    entity.Creator,
		CreateDate: entity.CreateDate,
		Updater:    entity.Updater,
		UpdateDate: entity.UpdateDate,
	}

	if !entity.LastConnectedAt.IsZero() {
		d.LastConnectedAt = &entity.LastConnectedAt
	}

	return d
}

// SelectCountByUserId 获取用户的设备数量
func (r *deviceRepo) SelectCountByUserId(ctx context.Context, userId int64) (int64, error) {
	count, err := r.data.db.Device.Query().
		Where(device.UserIDEQ(userId)).
		Count(ctx)
	return int64(count), err
}

// DeleteByUserId 删除用户的所有设备
func (r *deviceRepo) DeleteByUserId(ctx context.Context, userId int64) error {
	_, err := r.data.db.Device.Delete().
		Where(device.UserIDEQ(userId)).
		Exec(ctx)
	return err
}

// PageDevices 分页查询所有设备，支持按别名模糊查询
func (r *deviceRepo) PageDevices(ctx context.Context, keywords *string, page *kit.PageRequest) ([]*biz.Device, error) {
	query := r.data.db.Device.Query()

	// 如果提供了关键词，进行模糊查询
	if keywords != nil && *keywords != "" {
		query = query.Where(device.AliasContains(*keywords))
	}

	// 默认按mac_address降序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("mac_address")
		page.SetSortDesc()
	}

	applyPagination(query, page, device.Columns)

	devices, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.Device, len(devices))
	for i, device := range devices {
		result[i] = r.toBizDevice(device)
	}

	return result, nil
}

// TotalDevices 获取设备总数（支持按关键词过滤）
func (r *deviceRepo) TotalDevices(ctx context.Context, keywords *string) (int, error) {
	query := r.data.db.Device.Query()

	if keywords != nil && *keywords != "" {
		query = query.Where(device.AliasContains(*keywords))
	}

	return query.Count(ctx)
}

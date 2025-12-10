package data

import (
	"context"
	"time"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/data/ent"
	"github.com/weetime/agent-matrix/internal/data/ent/agentpluginmapping"
	"github.com/weetime/agent-matrix/internal/data/ent/ragdataset"
	"github.com/weetime/agent-matrix/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
)

type datasetRepo struct {
	data *Data
	log  *log.Helper
}

// NewDatasetRepo 初始化DatasetRepo
func NewDatasetRepo(data *Data, logger log.Logger) biz.DatasetRepo {
	return &datasetRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "agent-matrix-service/data/dataset")),
	}
}

// PageDatasets 分页查询知识库，支持按名称模糊查询和创建者过滤
func (r *datasetRepo) PageDatasets(ctx context.Context, name *string, creator int64, page *kit.PageRequest) ([]*biz.Dataset, int, error) {
	query := r.data.db.RagDataset.Query()

	// 添加查询条件
	if name != nil && *name != "" {
		query = query.Where(ragdataset.NameContains(*name))
	}
	if creator > 0 {
		query = query.Where(ragdataset.CreatorEQ(creator))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 默认按创建时间降序排序
	if page == nil || page.GetSortField() == "" {
		page = &kit.PageRequest{}
		page.SetSortField("created_at")
		page.SetSortDesc()
	}

	// 应用排序
	sortField := page.GetSortField()
	switch sortField {
	case "created_at":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ragdataset.FieldCreatedAt))
		} else {
			query = query.Order(ent.Asc(ragdataset.FieldCreatedAt))
		}
	case "id":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ragdataset.FieldID))
		} else {
			query = query.Order(ent.Asc(ragdataset.FieldID))
		}
	case "name":
		if page.GetSort() == kit.PageRequest_DESC {
			query = query.Order(ent.Desc(ragdataset.FieldName))
		} else {
			query = query.Order(ent.Asc(ragdataset.FieldName))
		}
	default:
		// 默认按创建时间降序
		query = query.Order(ent.Desc(ragdataset.FieldCreatedAt))
	}

	// 应用分页
	pageNo, _ := page.GetPageNo()
	pageSize := page.GetPageSize()
	if pageNo > 0 && pageSize > 0 {
		offset := (pageNo - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	datasets, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.Dataset, len(datasets))
	for i, d := range datasets {
		var bizDataset biz.Dataset
		if err := copier.Copy(&bizDataset, d); err != nil {
			return nil, 0, err
		}
		result[i] = &bizDataset
	}

	return result, total, nil
}

// GetByDatasetID 根据dataset_id查询
func (r *datasetRepo) GetByDatasetID(ctx context.Context, datasetId string) (*biz.Dataset, error) {
	dataset, err := r.data.db.RagDataset.Query().
		Where(ragdataset.DatasetIDEQ(datasetId)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizDataset biz.Dataset
	if err := copier.Copy(&bizDataset, dataset); err != nil {
		return nil, err
	}
	return &bizDataset, nil
}

// GetByID 根据id查询
func (r *datasetRepo) GetByID(ctx context.Context, id string) (*biz.Dataset, error) {
	dataset, err := r.data.db.RagDataset.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var bizDataset biz.Dataset
	if err := copier.Copy(&bizDataset, dataset); err != nil {
		return nil, err
	}
	return &bizDataset, nil
}

// Create 创建知识库
func (r *datasetRepo) Create(ctx context.Context, dataset *biz.Dataset) error {
	create := r.data.db.RagDataset.Create().
		SetID(dataset.ID).
		SetDatasetID(dataset.DatasetID).
		SetName(dataset.Name).
		SetStatus(dataset.Status)

	if dataset.RagModelID != "" {
		create.SetRagModelID(dataset.RagModelID)
	}
	if dataset.Description != "" {
		create.SetDescription(dataset.Description)
	}
	if dataset.Creator > 0 {
		create.SetCreator(dataset.Creator)
	}
	if !dataset.CreatedAt.IsZero() {
		create.SetCreatedAt(dataset.CreatedAt)
	} else {
		create.SetCreatedAt(time.Now())
	}
	if dataset.Updater > 0 {
		create.SetUpdater(dataset.Updater)
	}
	if !dataset.UpdatedAt.IsZero() {
		create.SetUpdatedAt(dataset.UpdatedAt)
	}

	_, err := create.Save(ctx)
	return err
}

// Update 更新知识库
func (r *datasetRepo) Update(ctx context.Context, dataset *biz.Dataset) error {
	update := r.data.db.RagDataset.UpdateOneID(dataset.ID)

	if dataset.DatasetID != "" {
		update.SetDatasetID(dataset.DatasetID)
	}
	if dataset.RagModelID != "" {
		update.SetRagModelID(dataset.RagModelID)
	}
	if dataset.Name != "" {
		update.SetName(dataset.Name)
	}
	if dataset.Description != "" {
		update.SetDescription(dataset.Description)
	}
	if dataset.Status >= 0 {
		update.SetStatus(dataset.Status)
	}
	if dataset.Updater > 0 {
		update.SetUpdater(dataset.Updater)
	}
	update.SetUpdatedAt(time.Now())

	return update.Exec(ctx)
}

// DeleteByID 删除知识库
func (r *datasetRepo) DeleteByID(ctx context.Context, id string) error {
	_, err := r.data.db.RagDataset.Delete().
		Where(ragdataset.IDEQ(id)).
		Exec(ctx)
	return err
}

// CheckDuplicateName 检查同名知识库
func (r *datasetRepo) CheckDuplicateName(ctx context.Context, name string, creator int64, excludeId *string) (bool, error) {
	query := r.data.db.RagDataset.Query().
		Where(
			ragdataset.NameEQ(name),
			ragdataset.CreatorEQ(creator),
		)

	if excludeId != nil && *excludeId != "" {
		query = query.Where(ragdataset.IDNEQ(*excludeId))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeletePluginMappingByKnowledgeBaseID 删除关联的插件映射
// 根据Java实现，知识库ID被用作plugin_id
func (r *datasetRepo) DeletePluginMappingByKnowledgeBaseID(ctx context.Context, knowledgeBaseId string) error {
	_, err := r.data.db.AgentPluginMapping.Delete().
		Where(agentpluginmapping.PluginIDEQ(knowledgeBaseId)).
		Exec(ctx)
	return err
}


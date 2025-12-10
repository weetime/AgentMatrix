package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/kit/cerrors"
	"github.com/weetime/agent-matrix/internal/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Dataset 知识库实体
type Dataset struct {
	ID          string
	DatasetID   string
	RagModelID  string
	Name        string
	Description string
	Status      int32
	Creator     int64
	CreatedAt   time.Time
	Updater     int64
	UpdatedAt   time.Time
}

// DatasetDTO 知识库DTO（用于返回给前端）
type DatasetDTO struct {
	ID            string
	DatasetID     string
	RagModelID    string
	Name          string
	Description   string
	Status        int32
	Creator       string // 格式化为字符串
	CreatedAt     string // 格式化为字符串
	Updater       string // 格式化为字符串
	UpdatedAt     string // 格式化为字符串
	DocumentCount int32  // 文档数量（从RAG API获取，暂时返回0）
}

// DatasetRepo 知识库数据访问接口
type DatasetRepo interface {
	PageDatasets(ctx context.Context, name *string, creator int64, page *kit.PageRequest) ([]*Dataset, int, error)
	GetByDatasetID(ctx context.Context, datasetId string) (*Dataset, error)
	GetByID(ctx context.Context, id string) (*Dataset, error)
	Create(ctx context.Context, dataset *Dataset) error
	Update(ctx context.Context, dataset *Dataset) error
	DeleteByID(ctx context.Context, id string) error
	CheckDuplicateName(ctx context.Context, name string, creator int64, excludeId *string) (bool, error)
	DeletePluginMappingByKnowledgeBaseID(ctx context.Context, knowledgeBaseId string) error
}

// DatasetUsecase 知识库业务逻辑
type DatasetUsecase struct {
	datasetRepo     DatasetRepo
	modelConfigRepo ModelConfigRepo
	log             *log.Helper
	handleError     *cerrors.HandleError
}

// NewDatasetUsecase 创建知识库用例
func NewDatasetUsecase(
	datasetRepo DatasetRepo,
	modelConfigRepo ModelConfigRepo,
	logger log.Logger,
) *DatasetUsecase {
	return &DatasetUsecase{
		datasetRepo:     datasetRepo,
		modelConfigRepo: modelConfigRepo,
		log:             log.NewHelper(log.With(logger, "module", "biz/dataset")),
		handleError:     cerrors.NewHandleError(logger),
	}
}

// PageDatasets 分页查询知识库列表
func (uc *DatasetUsecase) PageDatasets(ctx context.Context, name *string, creator int64, page *kit.PageRequest) ([]*DatasetDTO, int, error) {
	datasets, total, err := uc.datasetRepo.PageDatasets(ctx, name, creator, page)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	result := make([]*DatasetDTO, len(datasets))
	for i := range datasets {
		result[i] = uc.toDTO(datasets[i])
	}

	return result, total, nil
}

// GetDatasetByDatasetID 根据dataset_id获取知识库详情
func (uc *DatasetUsecase) GetDatasetByDatasetID(ctx context.Context, datasetId string) (*DatasetDTO, error) {
	dataset, err := uc.datasetRepo.GetByDatasetID(ctx, datasetId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if dataset == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("知识库不存在"))
	}

	return uc.toDTO(dataset), nil
}

// CreateDataset 创建知识库
func (uc *DatasetUsecase) CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*DatasetDTO, error) {
	// 获取当前用户ID
	currentUserId, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return nil, uc.handleError.ErrPermissionDenied(ctx, err)
	}

	// 检查同名知识库
	exists, err := uc.datasetRepo.CheckDuplicateName(ctx, req.Name, currentUserId, nil)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if exists {
		return nil, uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("知识库名称已存在"))
	}

	// 生成UUID作为id
	id := uuid.New().String()
	// 生成dataset_id（暂时使用id，后续需要调用RAG API创建数据集）
	datasetId := id

	// 创建知识库实体
	dataset := &Dataset{
		ID:          id,
		DatasetID:   datasetId,
		RagModelID:  req.RagModelID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Creator:     currentUserId,
		CreatedAt:   time.Now(),
		Updater:     currentUserId,
		UpdatedAt:   time.Now(),
	}

	// TODO: 调用RAG API创建数据集
	// 暂时跳过RAG API调用，直接保存到数据库
	// 后续实现RAG适配器后再补充

	// 保存到数据库
	if err := uc.datasetRepo.Create(ctx, dataset); err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return uc.toDTO(dataset), nil
}

// UpdateDataset 更新知识库
func (uc *DatasetUsecase) UpdateDataset(ctx context.Context, datasetId string, req *UpdateDatasetRequest) (*DatasetDTO, error) {
	// 获取当前用户ID
	currentUserId, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return nil, uc.handleError.ErrPermissionDenied(ctx, err)
	}

	// 获取现有知识库
	dataset, err := uc.datasetRepo.GetByDatasetID(ctx, datasetId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}
	if dataset == nil {
		return nil, uc.handleError.ErrNotFound(ctx, fmt.Errorf("知识库不存在"))
	}

	// 检查权限：用户只能更新自己创建的知识库
	if dataset.Creator != currentUserId {
		return nil, uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("无权限操作此知识库"))
	}

	// 检查同名知识库（排除当前记录）
	if req.Name != nil && *req.Name != "" {
		excludeId := dataset.ID
		exists, err := uc.datasetRepo.CheckDuplicateName(ctx, *req.Name, currentUserId, &excludeId)
		if err != nil {
			return nil, uc.handleError.ErrInternal(ctx, err)
		}
		if exists {
			return nil, uc.handleError.ErrAlreadyExists(ctx, fmt.Errorf("知识库名称已存在"))
		}
		dataset.Name = *req.Name
	}

	// 更新字段
	if req.Description != nil {
		dataset.Description = *req.Description
	}
	if req.RagModelID != nil {
		dataset.RagModelID = *req.RagModelID
	}
	if req.Status != nil {
		dataset.Status = *req.Status
	}
	dataset.Updater = currentUserId
	dataset.UpdatedAt = time.Now()

	// TODO: 调用RAG API更新数据集
	// 暂时跳过RAG API调用，直接更新数据库
	// 后续实现RAG适配器后再补充

	// 更新数据库
	if err := uc.datasetRepo.Update(ctx, dataset); err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return uc.toDTO(dataset), nil
}

// DeleteDataset 删除知识库
func (uc *DatasetUsecase) DeleteDataset(ctx context.Context, datasetId string) error {
	// 获取当前用户ID
	currentUserId, err := middleware.GetUserIdFromContext(ctx)
	if err != nil {
		return uc.handleError.ErrPermissionDenied(ctx, err)
	}

	// 获取现有知识库
	dataset, err := uc.datasetRepo.GetByDatasetID(ctx, datasetId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}
	if dataset == nil {
		return uc.handleError.ErrNotFound(ctx, fmt.Errorf("知识库不存在"))
	}

	// 检查权限：用户只能删除自己创建的知识库
	if dataset.Creator != currentUserId {
		return uc.handleError.ErrPermissionDenied(ctx, fmt.Errorf("无权限操作此知识库"))
	}

	// TODO: 调用RAG API删除数据集
	// 暂时跳过RAG API调用，直接删除数据库记录
	// 后续实现RAG适配器后再补充

	// 删除关联的插件映射
	if err := uc.datasetRepo.DeletePluginMappingByKnowledgeBaseID(ctx, dataset.ID); err != nil {
		uc.log.Warnf("Failed to delete plugin mappings for knowledge base %s: %v", dataset.ID, err)
		// 继续删除知识库，不因为插件映射删除失败而中断
	}

	// 删除知识库
	if err := uc.datasetRepo.DeleteByID(ctx, dataset.ID); err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// GetRAGModels 获取RAG模型列表
func (uc *DatasetUsecase) GetRAGModels(ctx context.Context) ([]map[string]interface{}, error) {
	models, err := uc.modelConfigRepo.GetRAGModelList(ctx)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		result[i] = map[string]interface{}{
			"id":        m.ID,
			"modelName": m.ModelName,
		}
	}

	return result, nil
}

// toDTO 转换为DTO
func (uc *DatasetUsecase) toDTO(dataset *Dataset) *DatasetDTO {
	creator := ""
	if dataset.Creator > 0 {
		creator = fmt.Sprintf("%d", dataset.Creator)
	}
	updater := ""
	if dataset.Updater > 0 {
		updater = fmt.Sprintf("%d", dataset.Updater)
	}

	return &DatasetDTO{
		ID:            dataset.ID,
		DatasetID:     dataset.DatasetID,
		RagModelID:    dataset.RagModelID,
		Name:          dataset.Name,
		Description:   dataset.Description,
		Status:        dataset.Status,
		Creator:       creator,
		CreatedAt:     dataset.CreatedAt.Format("2006-01-02 15:04:05"),
		Updater:       updater,
		UpdatedAt:     dataset.UpdatedAt.Format("2006-01-02 15:04:05"),
		DocumentCount: 0, // 暂时返回0，后续从RAG API获取
	}
}

// CreateDatasetRequest 创建知识库请求
type CreateDatasetRequest struct {
	RagModelID  string
	Name        string
	Description string
	Status      int32
}

// UpdateDatasetRequest 更新知识库请求
type UpdateDatasetRequest struct {
	Name        *string
	Description *string
	RagModelID  *string
	Status      *int32
}


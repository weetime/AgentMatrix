package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weetime/agent-matrix/internal/kit/cerrors"

	"github.com/go-kratos/kratos/v2/log"
)

// Document 文档实体
type Document struct {
	ID           string
	DocumentID   string
	DatasetID    string
	Name         string
	FileType     string
	FileSize     int64
	FilePath     string
	MetaFields   map[string]interface{}
	ChunkMethod  string
	ParserConfig map[string]interface{}
	Status       int32  // 0-未开始，1-进行中，2-已取消，3-已完成，4-失败
	Run          string // RAG状态字符串
	Creator      int64
	CreatedAt    time.Time
	Updater      int64
	UpdatedAt    time.Time
}

// DocumentDTO 文档DTO（用于返回给前端，与Java KnowledgeFilesDTO一致）
type DocumentDTO struct {
	ID           string // 格式化为字符串
	DocumentID   string
	DatasetID    string
	Name         string
	FileType     string
	FileSize     int64
	FilePath     string
	MetaFields   map[string]interface{}
	ChunkMethod  string
	ParserConfig map[string]interface{}
	Status       int32  // 0-未开始，1-进行中，2-已取消，3-已完成，4-失败
	Run          string // RAG状态字符串
	Creator      string // 格式化为字符串
	CreatedAt    string // 格式化为字符串
	Updater      string // 格式化为字符串
	UpdatedAt    string // 格式化为字符串
}

// RAGAdapter RAG适配器接口
type RAGAdapter interface {
	// GetDocumentList 分页查询文档列表
	GetDocumentList(ctx context.Context, datasetId string, queryParams map[string]interface{}, page, limit int) ([]*Document, int, error)
	// GetDocumentById 根据文档ID获取文档详情
	GetDocumentById(ctx context.Context, datasetId, documentId string) (*Document, error)
	// UploadDocument 上传文档到知识库
	UploadDocument(ctx context.Context, datasetId string, file []byte, fileName string, params map[string]interface{}) (*Document, error)
	// DeleteDocument 删除文档
	DeleteDocument(ctx context.Context, datasetId, documentId string) error
	// ParseDocuments 解析文档（切块）
	ParseDocuments(ctx context.Context, datasetId string, documentIds []string) (bool, error)
	// ListChunks 列出指定文档的切片
	ListChunks(ctx context.Context, datasetId, documentId string, params map[string]interface{}) (map[string]interface{}, error)
	// RetrievalTest 召回测试
	RetrievalTest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
}

// RAGAdapterFactory RAG适配器工厂
type RAGAdapterFactory interface {
	GetAdapter(adapterType string, config map[string]interface{}) (RAGAdapter, error)
}

// DocumentUsecase 文档业务逻辑
type DocumentUsecase struct {
	datasetUsecase    *DatasetUsecase
	modelConfigRepo   ModelConfigRepo
	ragAdapterFactory RAGAdapterFactory
	log               *log.Helper
	handleError       *cerrors.HandleError
}

// NewDocumentUsecase 创建文档用例
func NewDocumentUsecase(
	datasetUsecase *DatasetUsecase,
	modelConfigRepo ModelConfigRepo,
	ragAdapterFactory RAGAdapterFactory,
	logger log.Logger,
) *DocumentUsecase {
	return &DocumentUsecase{
		datasetUsecase:    datasetUsecase,
		modelConfigRepo:   modelConfigRepo,
		ragAdapterFactory: ragAdapterFactory,
		log:               log.NewHelper(log.With(logger, "module", "biz/document")),
		handleError:       cerrors.NewHandleError(logger),
	}
}

// PageDocuments 分页查询文档列表
func (uc *DocumentUsecase) PageDocuments(ctx context.Context, datasetId string, name *string, status *int32, page, limit int) ([]*DocumentDTO, int, error) {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	// 构建查询参数
	queryParams := make(map[string]interface{})
	if name != nil && *name != "" {
		queryParams["keywords"] = *name
	}
	if status != nil {
		queryParams["status"] = *status
	}

	// 调用适配器查询文档列表
	documents, total, err := adapter.GetDocumentList(ctx, datasetId, queryParams, page, limit)
	if err != nil {
		return nil, 0, uc.handleError.ErrInternal(ctx, err)
	}

	// 转换为DTO
	result := make([]*DocumentDTO, len(documents))
	for i := range documents {
		result[i] = uc.toDTO(documents[i])
	}

	return result, total, nil
}

// PageDocumentsByStatus 按状态查询文档列表
func (uc *DocumentUsecase) PageDocumentsByStatus(ctx context.Context, datasetId string, status int32, page, limit int) ([]*DocumentDTO, int, error) {
	return uc.PageDocuments(ctx, datasetId, nil, &status, page, limit)
}

// UploadDocument 上传文档
func (uc *DocumentUsecase) UploadDocument(ctx context.Context, datasetId string, file []byte, fileName string, name *string, chunkMethod *string, metaFields map[string]interface{}, parserConfig map[string]interface{}) (*DocumentDTO, error) {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 构建上传参数
	params := make(map[string]interface{})
	if name != nil && *name != "" {
		params["name"] = *name
	}
	if chunkMethod != nil && *chunkMethod != "" {
		params["chunk_method"] = *chunkMethod
	}
	if metaFields != nil {
		params["meta_fields"] = metaFields
	}
	if parserConfig != nil {
		params["parser_config"] = parserConfig
	}

	// 调用适配器上传文档
	document, err := adapter.UploadDocument(ctx, datasetId, file, fileName, params)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return uc.toDTO(document), nil
}

// DeleteDocument 删除文档
func (uc *DocumentUsecase) DeleteDocument(ctx context.Context, datasetId, documentId string) error {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	// 调用适配器删除文档
	err = adapter.DeleteDocument(ctx, datasetId, documentId)
	if err != nil {
		return uc.handleError.ErrInternal(ctx, err)
	}

	return nil
}

// ParseDocuments 解析文档（切块）
func (uc *DocumentUsecase) ParseDocuments(ctx context.Context, datasetId string, documentIds []string) (bool, error) {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return false, uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return false, uc.handleError.ErrInternal(ctx, err)
	}

	// 调用适配器解析文档
	success, err := adapter.ParseDocuments(ctx, datasetId, documentIds)
	if err != nil {
		return false, uc.handleError.ErrInternal(ctx, err)
	}

	return success, nil
}

// ListChunks 列出文档切片
func (uc *DocumentUsecase) ListChunks(ctx context.Context, datasetId, documentId string, keywords *string, page, pageSize int, chunkId *string) (map[string]interface{}, error) {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 构建查询参数
	params := make(map[string]interface{})
	if keywords != nil && *keywords != "" {
		params["keywords"] = *keywords
	}
	if page > 0 {
		params["page"] = page
	}
	if pageSize > 0 {
		params["page_size"] = pageSize
	}
	if chunkId != nil && *chunkId != "" {
		params["id"] = *chunkId
	}

	// 调用适配器列出切片
	result, err := adapter.ListChunks(ctx, datasetId, documentId, params)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return result, nil
}

// RetrievalTest 召回测试
func (uc *DocumentUsecase) RetrievalTest(ctx context.Context, datasetId string, question string, datasetIds []string, documentIds []string, page, pageSize int, similarityThreshold, vectorSimilarityWeight *float32, topK *int32, rerankId *string, keyword, highlight *bool, crossLanguages []string, metadataCondition map[string]interface{}) (map[string]interface{}, error) {
	// 获取RAG配置
	ragConfig, err := uc.getRAGConfig(ctx, datasetId)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 获取适配器
	adapter, err := uc.getAdapter(ragConfig)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	// 构建检索参数
	params := make(map[string]interface{})
	params["question"] = question
	if len(datasetIds) > 0 {
		params["datasetIds"] = datasetIds
	} else {
		// 如果未指定数据集ID，使用当前数据集
		params["datasetIds"] = []string{datasetId}
	}
	if len(documentIds) > 0 {
		params["documentIds"] = documentIds
	}
	if page > 0 {
		params["page"] = page
	}
	if pageSize > 0 {
		params["pageSize"] = pageSize
	}
	if similarityThreshold != nil {
		params["similarityThreshold"] = *similarityThreshold
	}
	if vectorSimilarityWeight != nil {
		params["vectorSimilarityWeight"] = *vectorSimilarityWeight
	}
	if topK != nil && *topK > 0 {
		params["topK"] = *topK
	}
	if rerankId != nil && *rerankId != "" {
		params["rerankId"] = *rerankId
	}
	if keyword != nil {
		params["keyword"] = *keyword
	}
	if highlight != nil {
		params["highlight"] = *highlight
	}
	if len(crossLanguages) > 0 {
		params["crossLanguages"] = crossLanguages
	}
	if metadataCondition != nil {
		params["metadataCondition"] = metadataCondition
	}

	// 调用适配器进行召回测试
	result, err := adapter.RetrievalTest(ctx, params)
	if err != nil {
		return nil, uc.handleError.ErrInternal(ctx, err)
	}

	return result, nil
}

// getRAGConfig 获取RAG配置
func (uc *DocumentUsecase) getRAGConfig(ctx context.Context, datasetId string) (map[string]interface{}, error) {
	// 获取知识库信息
	dataset, err := uc.datasetUsecase.GetDatasetByDatasetID(ctx, datasetId)
	if err != nil {
		return nil, fmt.Errorf("获取知识库失败: %w", err)
	}

	// 获取RAG模型配置
	modelConfig, err := uc.modelConfigRepo.GetModelConfigByID(ctx, dataset.RagModelID)
	if err != nil {
		return nil, fmt.Errorf("获取RAG模型配置失败: %w", err)
	}

	// 解析ConfigJSON
	var ragConfig map[string]interface{}
	if modelConfig.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(modelConfig.ConfigJSON), &ragConfig); err != nil {
			return nil, fmt.Errorf("解析RAG配置JSON失败: %w", err)
		}
	} else {
		ragConfig = make(map[string]interface{})
	}

	// 从配置中提取适配器类型，如果没有则默认使用ragflow
	adapterType, ok := ragConfig["type"].(string)
	if !ok || adapterType == "" {
		adapterType = "ragflow"
		ragConfig["type"] = adapterType
	}

	return ragConfig, nil
}

// getAdapter 获取RAG适配器
func (uc *DocumentUsecase) getAdapter(ragConfig map[string]interface{}) (RAGAdapter, error) {
	adapterType, ok := ragConfig["adapter_type"].(string)
	if !ok || adapterType == "" {
		adapterType = "ragflow" // 默认使用ragflow
	}

	adapter, err := uc.ragAdapterFactory.GetAdapter(adapterType, ragConfig)
	if err != nil {
		return nil, fmt.Errorf("获取RAG适配器失败: %w", err)
	}

	return adapter, nil
}

// toDTO 转换为DTO
func (uc *DocumentUsecase) toDTO(document *Document) *DocumentDTO {
	creator := ""
	if document.Creator > 0 {
		creator = fmt.Sprintf("%d", document.Creator)
	}
	updater := ""
	if document.Updater > 0 {
		updater = fmt.Sprintf("%d", document.Updater)
	}
	id := ""
	if document.ID != "" {
		id = document.ID
	} else {
		// 如果没有ID，使用DocumentID
		id = document.DocumentID
	}

	return &DocumentDTO{
		ID:           id,
		DocumentID:   document.DocumentID,
		DatasetID:    document.DatasetID,
		Name:         document.Name,
		FileType:     document.FileType,
		FileSize:     document.FileSize,
		FilePath:     document.FilePath,
		MetaFields:   document.MetaFields,
		ChunkMethod:  document.ChunkMethod,
		ParserConfig: document.ParserConfig,
		Status:       document.Status,
		Run:          document.Run,
		Creator:      creator,
		CreatedAt:    document.CreatedAt.Format("2006-01-02 15:04:05"),
		Updater:      updater,
		UpdatedAt:    document.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/weetime/agent-matrix/internal/biz"
	"github.com/weetime/agent-matrix/internal/kit"
	"github.com/weetime/agent-matrix/internal/middleware"
	pb "github.com/weetime/agent-matrix/protos/v1"

	"google.golang.org/protobuf/types/known/structpb"
)

type DatasetService struct {
	pb.UnimplementedDatasetServiceServer
	datasetUsecase  *biz.DatasetUsecase
	documentUsecase *biz.DocumentUsecase
}

func NewDatasetService(
	datasetUsecase *biz.DatasetUsecase,
	documentUsecase *biz.DocumentUsecase,
) *DatasetService {
	return &DatasetService{
		datasetUsecase:  datasetUsecase,
		documentUsecase: documentUsecase,
	}
}

// PageDatasets 分页查询知识库列表
func (s *DatasetService) PageDatasets(ctx context.Context, req *pb.PageDatasetsRequest) (*pb.Response, error) {
	// 解析过滤条件
	var name *string
	if req.Name != nil && req.Name.GetValue() != "" {
		n := req.Name.GetValue()
		name = &n
	}

	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 解析分页参数
	page := &kit.PageRequest{}
	pageNo := req.GetPage()
	if pageNo == 0 {
		pageNo = 1
	}
	pageSize := req.GetLimit()
	if pageSize == 0 {
		pageSize = kit.DEFAULT_PAGE_ZISE
	}
	page.SetPageNo(int(pageNo))
	page.SetPageSize(int(pageSize))
	page.SetSortDesc()
	page.SetSortField("created_at")

	// 查询列表
	list, total, err := s.datasetUsecase.PageDatasets(ctx, name, currentUserId, page)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 转换为VO列表
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"id":             item.ID,
			"dataset_id":     item.DatasetID,
			"rag_model_id":   item.RagModelID,
			"name":           item.Name,
			"description":    item.Description,
			"status":         item.Status,
			"creator":        item.Creator,
			"created_at":     item.CreatedAt,
			"updater":        item.Updater,
			"updated_at":     item.UpdatedAt,
			"document_count": item.DocumentCount,
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// GetDataset 获取知识库详情
func (s *DatasetService) GetDataset(ctx context.Context, req *pb.GetDatasetRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 查询知识库
	dataset, err := s.datasetUsecase.GetDatasetByDatasetID(ctx, req.GetDatasetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 检查权限：用户只能查看自己创建的知识库
	creator, _ := strconv.ParseInt(dataset.Creator, 10, 64)
	if creator != currentUserId {
		return &pb.Response{
			Code: 403,
			Msg:  "无权限",
		}, nil
	}

	// 构建响应数据
	vo := map[string]interface{}{
		"id":             dataset.ID,
		"dataset_id":     dataset.DatasetID,
		"rag_model_id":   dataset.RagModelID,
		"name":           dataset.Name,
		"description":    dataset.Description,
		"status":         dataset.Status,
		"creator":        dataset.Creator,
		"created_at":     dataset.CreatedAt,
		"updater":        dataset.Updater,
		"updated_at":     dataset.UpdatedAt,
		"document_count": dataset.DocumentCount,
	}

	dataStruct, err := structpb.NewStruct(vo)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// CreateDataset 创建知识库
func (s *DatasetService) CreateDataset(ctx context.Context, req *pb.CreateDatasetRequest) (*pb.Response, error) {
	// 构建请求
	createReq := &biz.CreateDatasetRequest{
		RagModelID:  req.GetRagModelId(),
		Name:        req.GetName(),
		Description: req.GetDescription().GetValue(),
		Status:      req.GetStatus(),
	}

	if createReq.Status == 0 {
		createReq.Status = 1 // 默认启用
	}

	// 创建知识库
	dataset, err := s.datasetUsecase.CreateDataset(ctx, createReq)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应数据
	vo := map[string]interface{}{
		"id":             dataset.ID,
		"dataset_id":     dataset.DatasetID,
		"rag_model_id":   dataset.RagModelID,
		"name":           dataset.Name,
		"description":    dataset.Description,
		"status":         dataset.Status,
		"creator":        dataset.Creator,
		"created_at":     dataset.CreatedAt,
		"updater":        dataset.Updater,
		"updated_at":     dataset.UpdatedAt,
		"document_count": dataset.DocumentCount,
	}

	dataStruct, err := structpb.NewStruct(vo)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// UpdateDataset 更新知识库
func (s *DatasetService) UpdateDataset(ctx context.Context, req *pb.UpdateDatasetRequest) (*pb.Response, error) {
	// 构建请求
	updateReq := &biz.UpdateDatasetRequest{}

	if req.Name != nil {
		name := req.Name.GetValue()
		updateReq.Name = &name
	}
	if req.Description != nil {
		desc := req.Description.GetValue()
		updateReq.Description = &desc
	}
	if req.RagModelId != nil {
		ragModelId := req.RagModelId.GetValue()
		updateReq.RagModelID = &ragModelId
	}
	if req.Status != nil {
		status := req.Status.GetValue()
		updateReq.Status = &status
	}

	// 更新知识库
	dataset, err := s.datasetUsecase.UpdateDataset(ctx, req.GetDatasetId(), updateReq)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应数据
	vo := map[string]interface{}{
		"id":             dataset.ID,
		"dataset_id":     dataset.DatasetID,
		"rag_model_id":   dataset.RagModelID,
		"name":           dataset.Name,
		"description":    dataset.Description,
		"status":         dataset.Status,
		"creator":        dataset.Creator,
		"created_at":     dataset.CreatedAt,
		"updater":        dataset.Updater,
		"updated_at":     dataset.UpdatedAt,
		"document_count": dataset.DocumentCount,
	}

	dataStruct, err := structpb.NewStruct(vo)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// DeleteDataset 删除知识库
func (s *DatasetService) DeleteDataset(ctx context.Context, req *pb.DeleteDatasetRequest) (*pb.Response, error) {
	// 删除知识库
	err := s.datasetUsecase.DeleteDataset(ctx, req.GetDatasetId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// BatchDeleteDatasets 批量删除知识库
func (s *DatasetService) BatchDeleteDatasets(ctx context.Context, req *pb.BatchDeleteDatasetsRequest) (*pb.Response, error) {
	ids := req.GetIds()
	if ids == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "知识库ID列表不能为空",
		}, nil
	}

	// 解析ID列表（逗号分隔）
	idArray := strings.Split(ids, ",")
	for _, datasetId := range idArray {
		datasetId = strings.TrimSpace(datasetId)
		if datasetId == "" {
			continue
		}

		// 删除知识库（内部会检查权限，使用dataset_id）
		err := s.datasetUsecase.DeleteDataset(ctx, datasetId)
		if err != nil {
			return &pb.Response{
				Code: 500,
				Msg:  fmt.Sprintf("删除知识库 %s 失败: %s", datasetId, err.Error()),
			}, nil
		}
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// GetRAGModels 获取RAG模型列表
func (s *DatasetService) GetRAGModels(ctx context.Context, req *pb.Empty) (*pb.Response, error) {
	// 查询RAG模型列表
	models, err := s.datasetUsecase.GetRAGModels(ctx)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应数据
	dataStruct, err := structpb.NewStruct(map[string]interface{}{
		"list": models,
	})
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// getCurrentUserID 从context获取当前用户ID（辅助函数）
func getCurrentUserID(ctx context.Context) (int64, error) {
	return middleware.GetUserIdFromContext(ctx)
}

// validateKnowledgeBasePermission 验证知识库权限
func (s *DatasetService) validateKnowledgeBasePermission(ctx context.Context, datasetId string, currentUserId int64) error {
	// 获取知识库信息
	dataset, err := s.datasetUsecase.GetDatasetByDatasetID(ctx, datasetId)
	if err != nil {
		return fmt.Errorf("知识库不存在")
	}

	// 检查权限：用户只能操作自己创建的知识库
	creator, _ := strconv.ParseInt(dataset.Creator, 10, 64)
	if creator != currentUserId {
		return fmt.Errorf("无权限操作此知识库")
	}

	return nil
}

// PageDocuments 分页查询文档列表
func (s *DatasetService) PageDocuments(ctx context.Context, req *pb.PageDocumentsRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 解析查询条件
	var name *string
	if req.Name != nil && req.Name.GetValue() != "" {
		n := req.Name.GetValue()
		name = &n
	}
	var status *int32
	if req.Status != nil {
		s := req.Status.GetValue()
		status = &s
	}

	// 解析分页参数
	page := req.GetPage()
	if page == 0 {
		page = 1
	}
	limit := req.GetLimit()
	if limit == 0 {
		limit = kit.DEFAULT_PAGE_ZISE
	}

	// 调用biz层
	list, total, err := s.documentUsecase.PageDocuments(ctx, req.GetDatasetId(), name, status, int(page), int(limit))
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应（注意ID格式化）
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"id":            item.ID,
			"document_id":   item.DocumentID,
			"dataset_id":    item.DatasetID,
			"name":          item.Name,
			"file_type":     item.FileType,
			"file_size":     item.FileSize,
			"file_path":     item.FilePath,
			"meta_fields":   item.MetaFields,
			"chunk_method":  item.ChunkMethod,
			"parser_config": item.ParserConfig,
			"status":        item.Status,
			"run":           item.Run,
			"creator":       item.Creator,
			"created_at":    item.CreatedAt,
			"updater":       item.Updater,
			"updated_at":    item.UpdatedAt,
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// PageDocumentsByStatus 按状态查询文档列表
func (s *DatasetService) PageDocumentsByStatus(ctx context.Context, req *pb.PageDocumentsByStatusRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 解析分页参数
	page := req.GetPage()
	if page == 0 {
		page = 1
	}
	limit := req.GetLimit()
	if limit == 0 {
		limit = kit.DEFAULT_PAGE_ZISE
	}

	// 调用biz层
	list, total, err := s.documentUsecase.PageDocumentsByStatus(ctx, req.GetDatasetId(), req.GetStatus(), int(page), int(limit))
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应（注意ID格式化）
	voList := make([]interface{}, 0, len(list))
	for _, item := range list {
		vo := map[string]interface{}{
			"id":            item.ID,
			"document_id":   item.DocumentID,
			"dataset_id":    item.DatasetID,
			"name":          item.Name,
			"file_type":     item.FileType,
			"file_size":     item.FileSize,
			"file_path":     item.FilePath,
			"meta_fields":   item.MetaFields,
			"chunk_method":  item.ChunkMethod,
			"parser_config": item.ParserConfig,
			"status":        item.Status,
			"run":           item.Run,
			"creator":       item.Creator,
			"created_at":    item.CreatedAt,
			"updater":       item.Updater,
			"updated_at":    item.UpdatedAt,
		}
		voList = append(voList, vo)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"total": int32(total),
		"list":  voList,
	}

	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// UploadDocument 上传文档
// 注意：文件上传需要特殊处理multipart/form-data，这里先实现基本结构
func (s *DatasetService) UploadDocument(ctx context.Context, req *pb.UploadDocumentRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// TODO: 文件上传需要从HTTP请求中获取multipart文件
	// 这里先返回错误，提示需要特殊处理
	return &pb.Response{
		Code: 400,
		Msg:  "文件上传需要通过HTTP multipart/form-data处理，请使用HTTP handler",
	}, nil
}

// DeleteDocument 删除文档
func (s *DatasetService) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 调用biz层删除文档
	err = s.documentUsecase.DeleteDocument(ctx, req.GetDatasetId(), req.GetDocumentId())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// ParseDocuments 解析文档（切块）
func (s *DatasetService) ParseDocuments(ctx context.Context, req *pb.ParseDocumentsRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 验证参数
	if len(req.GetDocumentIds()) == 0 {
		return &pb.Response{
			Code: 400,
			Msg:  "document_ids参数不能为空",
		}, nil
	}

	// 调用biz层解析文档
	success, err := s.documentUsecase.ParseDocuments(ctx, req.GetDatasetId(), req.GetDocumentIds())
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	if !success {
		return &pb.Response{
			Code: 500,
			Msg:  "文档解析失败，文档可能正在处理中",
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// ListChunks 列出文档切片
func (s *DatasetService) ListChunks(ctx context.Context, req *pb.ListChunksRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 解析查询参数
	var keywords *string
	if req.Keywords != nil && req.Keywords.GetValue() != "" {
		k := req.Keywords.GetValue()
		keywords = &k
	}
	var chunkId *string
	if req.Id != nil && req.Id.GetValue() != "" {
		c := req.Id.GetValue()
		chunkId = &c
	}

	// 解析分页参数
	page := req.GetPage()
	if page == 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 1024
	}

	// 调用biz层列出切片
	result, err := s.documentUsecase.ListChunks(ctx, req.GetDatasetId(), req.GetDocumentId(), keywords, int(page), int(pageSize), chunkId)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	// 构建响应数据
	dataStruct, err := structpb.NewStruct(result)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

// RetrievalTest 召回测试
func (s *DatasetService) RetrievalTest(ctx context.Context, req *pb.RetrievalTestRequest) (*pb.Response, error) {
	// 获取当前用户ID
	currentUserId, err := getCurrentUserID(ctx)
	if err != nil {
		return &pb.Response{
			Code: 401,
			Msg:  "未授权",
		}, nil
	}

	// 验证知识库权限
	err = s.validateKnowledgeBasePermission(ctx, req.GetDatasetId(), currentUserId)
	if err != nil {
		return &pb.Response{
			Code: 403,
			Msg:  err.Error(),
		}, nil
	}

	// 验证参数
	if req.GetQuestion() == "" {
		return &pb.Response{
			Code: 400,
			Msg:  "问题不能为空",
		}, nil
	}

	// 解析可选参数
	var similarityThreshold *float32
	if req.GetSimilarityThreshold() != 0 {
		st := req.GetSimilarityThreshold()
		similarityThreshold = &st
	}
	var vectorSimilarityWeight *float32
	if req.GetVectorSimilarityWeight() != 0 {
		vsw := req.GetVectorSimilarityWeight()
		vectorSimilarityWeight = &vsw
	}
	var topK *int32
	if req.GetTopK() != 0 {
		tk := req.GetTopK()
		topK = &tk
	}
	var rerankId *string
	if req.RerankId != nil && req.RerankId.GetValue() != "" {
		rid := req.RerankId.GetValue()
		rerankId = &rid
	}
	var keyword *bool
	if req.GetKeyword() {
		k := true
		keyword = &k
	}
	var highlight *bool
	if req.GetHighlight() {
		h := true
		highlight = &h
	}

	// 解析metadata_condition
	var metadataCondition map[string]interface{}
	if req.MetadataCondition != nil {
		metadataCondition = req.MetadataCondition.AsMap()
	}

	// 调用biz层进行召回测试
	result, err := s.documentUsecase.RetrievalTest(ctx, req.GetDatasetId(), req.GetQuestion(), req.GetDatasetIds(), req.GetDocumentIds(), int(req.GetPage()), int(req.GetPageSize()), similarityThreshold, vectorSimilarityWeight, topK, rerankId, keyword, highlight, req.GetCrossLanguages(), metadataCondition)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "召回测试失败: " + err.Error(),
		}, nil
	}

	// 构建响应数据
	dataStruct, err := structpb.NewStruct(result)
	if err != nil {
		return &pb.Response{
			Code: 500,
			Msg:  "构建响应数据失败: " + err.Error(),
		}, nil
	}

	return &pb.Response{
		Code: 0,
		Msg:  "success",
		Data: dataStruct,
	}, nil
}

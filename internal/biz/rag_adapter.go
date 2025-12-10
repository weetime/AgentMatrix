package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// RAGAdapterFactoryImpl RAG适配器工厂实现
type RAGAdapterFactoryImpl struct {
	log *log.Helper
}

// NewRAGAdapterFactory 创建RAG适配器工厂
func NewRAGAdapterFactory(logger log.Logger) RAGAdapterFactory {
	return &RAGAdapterFactoryImpl{
		log: log.NewHelper(log.With(logger, "module", "biz/rag_adapter")),
	}
}

// GetAdapter 获取RAG适配器
func (f *RAGAdapterFactoryImpl) GetAdapter(adapterType string, config map[string]interface{}) (RAGAdapter, error) {
	switch adapterType {
	case "ragflow":
		adapter := &RAGFlowAdapter{
			config: config,
			log:    f.log,
		}
		// 初始化适配器
		if err := adapter.validateConfig(config); err != nil {
			return nil, fmt.Errorf("RAG配置验证失败: %w", err)
		}
		return adapter, nil
	default:
		return nil, fmt.Errorf("不支持的适配器类型: %s", adapterType)
	}
}

// RAGFlowAdapter RAGFlow适配器实现
type RAGFlowAdapter struct {
	config map[string]interface{}
	log    *log.Helper
}

// GetDocumentList 分页查询文档列表
func (a *RAGFlowAdapter) GetDocumentList(ctx context.Context, datasetId string, queryParams map[string]interface{}, page, limit int) ([]*Document, int, error) {
	// TODO: 实现RAGFlow API调用
	// 这里先返回空结果，后续需要实现HTTP客户端调用RAGFlow API
	a.log.Warnf("GetDocumentList not implemented yet for datasetId: %s", datasetId)
	return []*Document{}, 0, fmt.Errorf("RAGFlow适配器未完全实现")
}

// GetDocumentById 根据文档ID获取文档详情
func (a *RAGFlowAdapter) GetDocumentById(ctx context.Context, datasetId, documentId string) (*Document, error) {
	// TODO: 实现RAGFlow API调用
	a.log.Warnf("GetDocumentById not implemented yet for datasetId: %s, documentId: %s", datasetId, documentId)
	return nil, fmt.Errorf("RAGFlow适配器未完全实现")
}

// UploadDocument 上传文档到知识库
func (a *RAGFlowAdapter) UploadDocument(ctx context.Context, datasetId string, file []byte, fileName string, params map[string]interface{}) (*Document, error) {
	// TODO: 实现RAGFlow API调用，使用multipart/form-data上传文件
	a.log.Warnf("UploadDocument not implemented yet for datasetId: %s, fileName: %s", datasetId, fileName)
	return nil, fmt.Errorf("RAGFlow适配器未完全实现")
}

// DeleteDocument 删除文档
func (a *RAGFlowAdapter) DeleteDocument(ctx context.Context, datasetId, documentId string) error {
	// TODO: 实现RAGFlow API调用
	a.log.Warnf("DeleteDocument not implemented yet for datasetId: %s, documentId: %s", datasetId, documentId)
	return fmt.Errorf("RAGFlow适配器未完全实现")
}

// ParseDocuments 解析文档（切块）
func (a *RAGFlowAdapter) ParseDocuments(ctx context.Context, datasetId string, documentIds []string) (bool, error) {
	// TODO: 实现RAGFlow API调用
	a.log.Warnf("ParseDocuments not implemented yet for datasetId: %s", datasetId)
	return false, fmt.Errorf("RAGFlow适配器未完全实现")
}

// ListChunks 列出指定文档的切片
func (a *RAGFlowAdapter) ListChunks(ctx context.Context, datasetId, documentId string, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: 实现RAGFlow API调用
	a.log.Warnf("ListChunks not implemented yet for datasetId: %s, documentId: %s", datasetId, documentId)
	return nil, fmt.Errorf("RAGFlow适配器未完全实现")
}

// RetrievalTest 召回测试
func (a *RAGFlowAdapter) RetrievalTest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: 实现RAGFlow API调用
	a.log.Warnf("RetrievalTest not implemented yet")
	return nil, fmt.Errorf("RAGFlow适配器未完全实现")
}

// validateConfig 验证配置
func (a *RAGFlowAdapter) validateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("RAG配置不能为空")
	}

	baseUrl, ok := config["base_url"].(string)
	if !ok || baseUrl == "" {
		return fmt.Errorf("RAG配置缺少base_url")
	}

	if baseUrl[:7] != "http://" && baseUrl[:8] != "https://" {
		return fmt.Errorf("RAG配置base_url格式无效")
	}

	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("RAG配置缺少api_key")
	}

	return nil
}

// getBaseURL 获取RAGFlow基础URL
func (a *RAGFlowAdapter) getBaseURL() string {
	if baseURL, ok := a.config["base_url"].(string); ok {
		return baseURL
	}
	return ""
}

// getAPIKey 获取RAGFlow API Key
func (a *RAGFlowAdapter) getAPIKey() string {
	if apiKey, ok := a.config["api_key"].(string); ok {
		return apiKey
	}
	return ""
}

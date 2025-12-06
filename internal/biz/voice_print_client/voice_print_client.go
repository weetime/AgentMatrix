package voice_print_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// IdentifyVoicePrintResponse 声纹识别响应
type IdentifyVoicePrintResponse struct {
	SpeakerID string  `json:"speaker_id"`
	Score     float64 `json:"score"`
}

// VoicePrintClient 声纹服务HTTP客户端
type VoicePrintClient struct {
	httpClient *http.Client
	log        *log.Helper
}

// NewVoicePrintClient 创建声纹服务客户端
func NewVoicePrintClient(logger log.Logger) *VoicePrintClient {
	return &VoicePrintClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log.NewHelper(log.With(logger, "module", "voice_print_client")),
	}
}

// getVoicePrintURI 获取声纹接口URI对象
func (c *VoicePrintClient) getVoicePrintURI(voicePrintURL string) (*url.URL, error) {
	uri, err := url.Parse(voicePrintURL)
	if err != nil {
		return nil, fmt.Errorf("声纹接口地址格式错误: %w", err)
	}
	return uri, nil
}

// getBaseUrl 获取声纹地址基础路径
func (c *VoicePrintClient) getBaseUrl(uri *url.URL) string {
	scheme := uri.Scheme
	host := uri.Host
	port := uri.Port()
	if port == "" {
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

// getAuthorization 获取验证Authorization
func (c *VoicePrintClient) getAuthorization(uri *url.URL) (string, error) {
	query := uri.RawQuery
	if query == "" {
		return "", fmt.Errorf("声纹接口地址缺少key参数")
	}
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("解析query参数失败: %w", err)
	}
	key := values.Get("key")
	if key == "" {
		return "", fmt.Errorf("声纹接口地址缺少key参数")
	}
	return fmt.Sprintf("Bearer %s", key), nil
}

// RegisterVoicePrint 注册声纹
func (c *VoicePrintClient) RegisterVoicePrint(voicePrintURL, id string, audioData []byte) error {
	uri, err := c.getVoicePrintURI(voicePrintURL)
	if err != nil {
		return err
	}

	baseUrl := c.getBaseUrl(uri)
	requestUrl := fmt.Sprintf("%s/voiceprint/register", baseUrl)

	// 创建multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加speaker_id字段
	if err := writer.WriteField("speaker_id", id); err != nil {
		return fmt.Errorf("写入speaker_id失败: %w", err)
	}

	// 添加文件字段
	part, err := writer.CreateFormFile("file", "VoicePrint.WAV")
	if err != nil {
		return fmt.Errorf("创建文件字段失败: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return fmt.Errorf("写入音频数据失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭writer失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", requestUrl, &requestBody)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	auth, err := c.getAuthorization(uri)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("声纹注册请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("声纹注册失败,请求路径：%s, 状态码：%d", requestUrl, resp.StatusCode)
		return fmt.Errorf("声纹注册请求失败，状态码：%d", resp.StatusCode)
	}

	// 检查响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "true") {
		c.log.Errorf("声纹注册失败,请求处理失败内容：%s", bodyStr)
		return fmt.Errorf("声纹注册处理失败")
	}

	return nil
}

// CancelVoicePrint 注销声纹
func (c *VoicePrintClient) CancelVoicePrint(voicePrintURL, voicePrintId string) error {
	uri, err := c.getVoicePrintURI(voicePrintURL)
	if err != nil {
		return err
	}

	baseUrl := c.getBaseUrl(uri)
	requestUrl := fmt.Sprintf("%s/voiceprint/%s", baseUrl, voicePrintId)

	// 创建请求
	req, err := http.NewRequest("DELETE", requestUrl, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	auth, err := c.getAuthorization(uri)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("声纹注销请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("声纹注销失败,请求路径：%s, 状态码：%d", requestUrl, resp.StatusCode)
		return fmt.Errorf("声纹注销请求失败，状态码：%d", resp.StatusCode)
	}

	// 检查响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "true") {
		c.log.Errorf("声纹注销失败,请求处理失败内容：%s", bodyStr)
		return fmt.Errorf("声纹注销处理失败")
	}

	return nil
}

// IdentifyVoicePrint 识别声纹
func (c *VoicePrintClient) IdentifyVoicePrint(voicePrintURL string, speakerIds []string, audioData []byte) (*IdentifyVoicePrintResponse, error) {
	uri, err := c.getVoicePrintURI(voicePrintURL)
	if err != nil {
		return nil, err
	}

	baseUrl := c.getBaseUrl(uri)
	requestUrl := fmt.Sprintf("%s/voiceprint/identify", baseUrl)

	// 创建multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加speaker_ids字段
	speakerIdsStr := strings.Join(speakerIds, ",")
	if err := writer.WriteField("speaker_ids", speakerIdsStr); err != nil {
		return nil, fmt.Errorf("写入speaker_ids失败: %w", err)
	}

	// 添加文件字段
	part, err := writer.CreateFormFile("file", "VoicePrint.WAV")
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, fmt.Errorf("写入音频数据失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭writer失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", requestUrl, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	auth, err := c.getAuthorization(uri)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("声纹识别请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("声纹识别请求失败,请求路径：%s, 状态码：%d", requestUrl, resp.StatusCode)
		return nil, fmt.Errorf("声纹识别请求失败，状态码：%d", resp.StatusCode)
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if len(body) == 0 {
		return nil, nil
	}

	var response IdentifyVoicePrintResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

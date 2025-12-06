package kit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
)

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	conn        *websocket.Conn
	url         string
	timeout     time.Duration
	maxDuration time.Duration
	bufferSize  int
	responses   []string
	mu          sync.Mutex
	log         *log.Helper
	ctx         context.Context
	cancel      context.CancelFunc
}

// WebSocketClientConfig WebSocket客户端配置
type WebSocketClientConfig struct {
	URL            string
	ConnectTimeout time.Duration
	MaxDuration    time.Duration
	BufferSize     int
	Logger         *log.Helper
}

// NewWebSocketClient 创建WebSocket客户端
func NewWebSocketClient(config *WebSocketClientConfig) *WebSocketClient {
	ctx, cancel := context.WithTimeout(context.Background(), config.MaxDuration)
	var logger *log.Helper
	if config.Logger != nil {
		logger = config.Logger
	} else {
		logger = log.NewHelper(log.DefaultLogger)
	}
	return &WebSocketClient{
		url:         config.URL,
		timeout:     config.ConnectTimeout,
		maxDuration: config.MaxDuration,
		bufferSize:  config.BufferSize,
		responses:   make([]string, 0),
		log:         logger,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Connect 连接到WebSocket服务器
func (c *WebSocketClient) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("解析URL失败: %w", err)
	}

	// 设置连接超时
	dialer := websocket.Dialer{
		HandshakeTimeout: c.timeout,
		ReadBufferSize:  c.bufferSize,
		WriteBufferSize: c.bufferSize,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("WebSocket连接失败: %w", err)
	}

	c.conn = conn

	// 启动消息接收goroutine
	go c.receiveMessages()

	return nil
}

// receiveMessages 接收消息
func (c *WebSocketClient) receiveMessages() {
	defer c.cancel()
	
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.log.Warnf("WebSocket读取错误: %v", err)
				}
				return
			}

			c.mu.Lock()
			c.responses = append(c.responses, string(message))
			c.mu.Unlock()
		}
	}
}

// SendText 发送文本消息
func (c *WebSocketClient) SendText(message string) error {
	if c.conn == nil {
		return fmt.Errorf("WebSocket未连接")
	}

	writeDeadline := time.Now().Add(5 * time.Second)
	if err := c.conn.SetWriteDeadline(writeDeadline); err != nil {
		return fmt.Errorf("设置写入超时失败: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// ListenForResponse 监听响应（带过滤条件）
// filter: 过滤函数，返回true表示匹配
// 返回匹配的响应列表
func (c *WebSocketClient) ListenForResponse(filter func(string) bool) ([]string, error) {
	deadline := time.Now().Add(10 * time.Second)
	var matched []string

	for {
		select {
		case <-c.ctx.Done():
			return matched, nil
		case <-time.After(time.Until(deadline)):
			// 超时，返回已匹配的响应
			c.mu.Lock()
			for _, resp := range c.responses {
				if filter(resp) {
					matched = append(matched, resp)
				}
			}
			c.mu.Unlock()
			return matched, nil
		default:
			time.Sleep(100 * time.Millisecond)
			c.mu.Lock()
			for _, resp := range c.responses {
				if filter(resp) {
					matched = append(matched, resp)
				}
			}
			c.mu.Unlock()
			if len(matched) > 0 {
				return matched, nil
			}
		}
	}
}

// ListenForResponseWithoutClose 监听响应（不关闭连接）
func (c *WebSocketClient) ListenForResponseWithoutClose(filter func(string) bool) []string {
	var matched []string
	deadline := time.Now().Add(10 * time.Second)

	for {
		select {
		case <-c.ctx.Done():
			return matched
		case <-time.After(time.Until(deadline)):
			// 超时，返回已匹配的响应
			c.mu.Lock()
			for _, resp := range c.responses {
				if filter(resp) {
					matched = append(matched, resp)
				}
			}
			c.mu.Unlock()
			return matched
		default:
			time.Sleep(100 * time.Millisecond)
			c.mu.Lock()
			for _, resp := range c.responses {
				if filter(resp) {
					matched = append(matched, resp)
				}
			}
			c.mu.Unlock()
			if len(matched) > 0 {
				return matched
			}
		}
	}
}

// Close 关闭连接
func (c *WebSocketClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ParseJsonResponse 解析JSON响应
func ParseJsonResponse(response string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}
	return result, nil
}


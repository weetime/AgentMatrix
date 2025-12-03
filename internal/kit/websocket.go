package kit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
)

const (
	ReceiveScopeUser      = "user"
	ReceiveScopeNamespace = "namespace"
)

const (
	StatusMessageType   = "statusMsg"
	PingMessageType     = "ping"
	PongMessageType     = "pong"
	HeartbeatInterval   = 30 * time.Second // 心跳间隔
	MaxMissedHeartbeats = 3                // 允许的最大心跳丢失次数
)

var ws = &WebSocket{
	upgrader: &websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源的连接
		},
	},
	// 节点服务连接池
	nodeConnections: make(map[string][]*ConnInfo), // Node服务连接池，key为nodeID
	connectionMux:   &sync.RWMutex{},              // 用于保护连接池的并发访问
	done:            make(chan struct{}),
}

// GetWebSocket returns the singleton WebSocket instance
func GetWebSocket() *WebSocket {
	return ws
}

// ConnInfo 保存连接相关信息
type ConnInfo struct {
	Conn             *websocket.Conn
	ID               string     // 节点ID
	ConnTime         time.Time  // 连接建立时间
	LastPongTime     time.Time  // 最后一次Pong响应时间
	MissedHeartbeats int        // 未响应的心跳次数
	mu               sync.Mutex // 保护ConnInfo的并发访问
}

// WebSocketMessage 定义WebSocket消息的基本结构
type WebSocketMessage struct {
	Type         string    `json:"type"`         // 消息类型
	Payload      Payload   `json:"payload"`      // 消息内容
	ReceiverName string    `json:"receiverName"` // 接收者名称
	ReceiveScope string    `json:"receiveScope"` // 接收者类型
	Timestamp    time.Time `json:"timestamp"`    // 时间戳
}

type Payload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// WebSocket 处理WebSocket连接和消息
type WebSocket struct {
	upgrader        *websocket.Upgrader
	nodeConnections map[string][]*ConnInfo // Node服务连接池，key为nodeID
	connectionMux   *sync.RWMutex          // 用于保护连接池的并发访问
	done            chan struct{}          // 用于关闭心跳检测goroutine
}

// InitWebSocket 初始化WebSocket并启动心跳检测
func InitWebSocket() {
	ws.startHeartbeatChecker()
}

// 关闭WebSocket并清理资源
func (s *WebSocket) Shutdown() {
	close(s.done)

	// 关闭所有连接
	s.connectionMux.Lock()
	defer s.connectionMux.Unlock()

	// 关闭节点连接
	for _, conns := range s.nodeConnections {
		for _, ci := range conns {
			ci.Conn.Close()
		}
	}

	// 清空连接
	s.nodeConnections = make(map[string][]*ConnInfo)
}

// 启动心跳检测器
func (s *WebSocket) startHeartbeatChecker() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkHeartbeats()
		case <-s.done:
			return
		}
	}
}

// 检查所有连接的心跳状态
func (s *WebSocket) checkHeartbeats() {
	s.connectionMux.Lock()
	defer s.connectionMux.Unlock()

	// 检查节点连接
	for nodeID, conns := range s.nodeConnections {
		var activeConns []*ConnInfo

		for _, ci := range conns {
			ci.mu.Lock()

			// 发送心跳
			err := ci.Conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
			if err != nil {
				// 连接可能已断开
				log.Infof("心跳发送失败，关闭连接: %v, 节点ID: %s", err, ci.ID)
				ci.Conn.Close()
				ci.mu.Unlock()
				continue
			}

			// 增加未响应心跳计数
			ci.MissedHeartbeats++

			// 如果超过最大未响应次数，关闭连接
			if ci.MissedHeartbeats > MaxMissedHeartbeats {
				log.Infof("连接 %d 次未响应心跳，关闭连接, 节点ID: %s", MaxMissedHeartbeats, ci.ID)
				ci.Conn.Close()
				ci.mu.Unlock()
				continue
			}

			ci.mu.Unlock()
			activeConns = append(activeConns, ci)
		}

		// 更新活跃连接
		if len(activeConns) == 0 {
			delete(s.nodeConnections, nodeID)
			log.Infof("节点 %s 的所有连接已关闭", nodeID)
		} else {
			s.nodeConnections[nodeID] = activeConns
		}
	}
}

// NodeWebSocketHandler 处理来自Node服务的WebSocket连接请求
func (s *WebSocket) NodeWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 从URL查询参数获取节点ID
	nodeID := r.URL.Query().Get("identifier")
	if nodeID == "" {
		log.Errorf("缺少identifier参数")
		conn.Close()
		return
	}

	// 注册新的节点连接
	connInfo := s.registerNodeClient(conn, nodeID)

	// 设置pong处理函数
	conn.SetPongHandler(func(string) error {
		connInfo.mu.Lock()
		connInfo.LastPongTime = time.Now()
		connInfo.MissedHeartbeats = 0 // 重置未响应计数
		connInfo.mu.Unlock()
		return nil
	})

	// 确保连接关闭时取消注册
	defer func() {
		s.unregisterNodeClient(nodeID, conn)
		conn.Close()
	}()

	// 处理接收到的消息
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("WebSocket读取错误: %v", err)
			}
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Errorf("消息解析失败: %v", err)
			continue
		}
		// fmt.Println("wsMsg", wsMsg)
		if err := s.handleNodeMessage(wsMsg, conn, nodeID); err != nil {
			log.Errorf("节点消息处理失败: %v", err)
		}
	}
}

// 处理来自Node服务的消息
func (s *WebSocket) handleNodeMessage(msg WebSocketMessage, conn *websocket.Conn, nodeID string) error {
	switch msg.Type {
	case PingMessageType:
		return s.handlePing(conn)
	// 节点特有的消息类型可以在这里处理
	default:
		log.Warnf("未知节点消息类型: %s, 节点ID: %s", msg.Type, nodeID)
		return nil
	}
}

// 注册新的Node服务客户端连接
func (s *WebSocket) registerNodeClient(conn *websocket.Conn, nodeID string) *ConnInfo {
	s.connectionMux.Lock()
	defer s.connectionMux.Unlock()

	now := time.Now()
	connInfo := &ConnInfo{
		Conn:             conn,
		ID:               nodeID,
		ConnTime:         now,
		LastPongTime:     now,
		MissedHeartbeats: 0,
	}

	s.nodeConnections[nodeID] = append(s.nodeConnections[nodeID], connInfo)

	log.Infof("新节点连接: %s, 连接总数: %d", nodeID, len(s.nodeConnections[nodeID]))

	return connInfo
}

// 注销节点客户端连接
func (s *WebSocket) unregisterNodeClient(nodeID string, conn *websocket.Conn) {
	s.connectionMux.Lock()
	defer s.connectionMux.Unlock()

	conns, exists := s.nodeConnections[nodeID]
	if !exists {
		return
	}

	// 查找并移除特定连接
	for i, ci := range conns {
		if ci.Conn == conn {
			// 从切片中移除该连接
			s.nodeConnections[nodeID] = append(conns[:i], conns[i+1:]...)
			log.Infof("节点连接断开: %s, 剩余连接: %d", nodeID, len(s.nodeConnections[nodeID]))

			// 如果该节点没有剩余连接，删除该节点条目
			if len(s.nodeConnections[nodeID]) == 0 {
				delete(s.nodeConnections, nodeID)
				log.Infof("节点 %s 所有连接已关闭", nodeID)
			}
			break
		}
	}
}

// 广播消息给所有节点服务
func (s *WebSocket) BroadcastToAllNodes(message WebSocketMessage) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Errorf("节点广播消息序列化失败: %v", err)
		return
	}

	s.connectionMux.RLock()
	defer s.connectionMux.RUnlock()

	// 广播给所有节点
	for nodeID, conns := range s.nodeConnections {
		for _, ci := range conns {
			err := ci.Conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				log.Errorf("向节点 %s 广播消息失败: %v", nodeID, err)
				// 错误处理在心跳检测中进行
			}
			fmt.Println("广播消息给节点", string(msgBytes), nodeID)
		}
	}
}

// 广播消息给特定节点的所有连接
func (s *WebSocket) BroadcastToNode(nodeID string, message WebSocketMessage) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Errorf("节点消息序列化失败: %v", err)
		return
	}

	s.connectionMux.RLock()
	defer s.connectionMux.RUnlock()

	conns, exists := s.nodeConnections[nodeID]
	if !exists {
		log.Warnf("节点 %s 没有活跃连接", nodeID)
		return
	}

	for _, ci := range conns {
		err := ci.Conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			log.Errorf("向节点 %s 发送消息失败: %v", nodeID, err)
			// 错误处理在心跳检测中进行
		}
	}
}

// handlePing 处理ping消息
func (s *WebSocket) handlePing(conn *websocket.Conn) error {
	// 获取远程地址
	remoteAddr := conn.RemoteAddr().String()

	response := WebSocketMessage{
		Type: PongMessageType,
		Payload: Payload{
			Message: "nova pong from " + remoteAddr,
			Status:  "success",
		},
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	// todo fy 更新注册器中uuid的时间戳，超过10min没有更新，则删除
	return conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// 获取当前节点连接总数
func (s *WebSocket) GetTotalConnections() int {
	s.connectionMux.RLock()
	defer s.connectionMux.RUnlock()

	total := 0

	// 统计节点连接
	for _, conns := range s.nodeConnections {
		total += len(conns)
	}

	return total
}

// 获取当前活跃节点数
func (s *WebSocket) GetActiveNodeCount() int {
	s.connectionMux.RLock()
	defer s.connectionMux.RUnlock()

	return len(s.nodeConnections)
}

// 获取所有活跃节点ID
func (s *WebSocket) GetActiveNodeIDs() []string {
	s.connectionMux.RLock()
	defer s.connectionMux.RUnlock()

	nodeIDs := make([]string, 0, len(s.nodeConnections))
	for nodeID := range s.nodeConnections {
		nodeIDs = append(nodeIDs, nodeID)
	}

	return nodeIDs
}

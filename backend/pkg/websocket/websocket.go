package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境中应该更严格）
		return true
	},
}

// 消息类型
type MessageType string

const (
	MessageTypeDeviceStatusUpdate MessageType = "device_status_update"
	MessageTypeTemperatureData    MessageType = "temperature_data"
	MessageTypeBreakerData        MessageType = "breaker_data"
	MessageTypeServerData         MessageType = "server_data"
	MessageTypeAlarmTriggered     MessageType = "alarm_triggered"
	MessageTypeAIControlExecuted  MessageType = "ai_control_executed"
	MessageTypePing               MessageType = "ping"
	MessageTypePong               MessageType = "pong"
)

// WebSocket消息结构
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// 客户端连接
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan Message
	Hub    *Hub
	UserID string // 用户ID，用于权限控制
}

// WebSocket Hub
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket客户端已连接: %s", client.ID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket客户端已断开: %s", client.ID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// 广播消息
func (h *Hub) BroadcastMessage(msgType MessageType, data interface{}) {
	message := Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	h.broadcast <- message
}

// 获取连接的客户端数量
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// 发送消息给特定客户端
func (h *Hub) SendToClient(clientID string, msgType MessageType, data interface{}) bool {
	message := Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.ID == clientID {
			select {
			case client.Send <- message:
				return true
			default:
				return false
			}
		}
	}
	return false
}

// 发送消息给特定用户的所有客户端
func (h *Hub) SendToUser(userID string, msgType MessageType, data interface{}) int {
	message := Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	sent := 0
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
				sent++
			default:
				// 客户端发送缓冲区满，跳过
			}
		}
	}
	return sent
}

// 获取所有连接的客户端信息
func (h *Hub) GetClientsInfo() []map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]map[string]interface{}, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, map[string]interface{}{
			"id":          client.ID,
			"user_id":     client.UserID,
			"remote_addr": client.Conn.RemoteAddr().String(),
		})
	}
	return clients
}

// 心跳检测
func (h *Hub) StartHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.BroadcastMessage(MessageTypePing, map[string]interface{}{
					"timestamp": time.Now().Unix(),
				})
			}
		}
	}()
}

// 实时数据推送方法
func (h *Hub) PushTemperatureData(data interface{}) {
	h.BroadcastMessage(MessageTypeTemperatureData, data)
}

func (h *Hub) PushServerData(data interface{}) {
	h.BroadcastMessage(MessageTypeServerData, data)
}

func (h *Hub) PushBreakerData(data interface{}) {
	h.BroadcastMessage(MessageTypeBreakerData, data)
}

func (h *Hub) PushAlarmTriggered(data interface{}) {
	h.BroadcastMessage(MessageTypeAlarmTriggered, data)
}

func (h *Hub) PushAIControlExecuted(data interface{}) {
	h.BroadcastMessage(MessageTypeAIControlExecuted, data)
}

func (h *Hub) PushDeviceStatusUpdate(data interface{}) {
	h.BroadcastMessage(MessageTypeDeviceStatusUpdate, data)
}

// 处理WebSocket连接
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 从查询参数或JWT token中获取用户ID
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		ID:     generateClientID(),
		Conn:   conn,
		Send:   make(chan Message, 256),
		Hub:    h,
		UserID: userID,
	}

	client.Hub.register <- client

	// 启动goroutines处理读写
	go client.writePump()
	go client.readPump()
}

// 读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("消息解析错误: %v", err)
			continue
		}

		// 处理ping消息
		if message.Type == MessageTypePing {
			pongMessage := Message{
				Type:      MessageTypePong,
				Data:      map[string]interface{}{"timestamp": time.Now().Unix()},
				Timestamp: time.Now().Unix(),
			}
			select {
			case c.Send <- pongMessage:
			default:
				close(c.Send)
				return
			}
		}
	}
}

// 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("消息序列化错误: %v", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("WebSocket写入错误: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// 数据处理器类型
type DataHandler func(data interface{})

// 数据处理器注册表
var dataHandlers = make(map[string][]DataHandler)
var handlerMutex sync.RWMutex

// 全局Hub实例
var GlobalHub *Hub

// 初始化WebSocket Hub
func InitWebSocketHub() {
	GlobalHub = NewHub()
	go GlobalHub.Run()
	log.Println("WebSocket Hub已启动")
}

// RegisterDataHandler 注册数据处理器
func RegisterDataHandler(dataType string, handler DataHandler) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	if dataHandlers[dataType] == nil {
		dataHandlers[dataType] = make([]DataHandler, 0)
	}
	dataHandlers[dataType] = append(dataHandlers[dataType], handler)
	log.Printf("注册数据处理器: %s", dataType)
}

// notifyDataHandlers 通知数据处理器
func notifyDataHandlers(dataType string, data interface{}) {
	handlerMutex.RLock()
	handlers := dataHandlers[dataType]
	handlerMutex.RUnlock()

	for _, handler := range handlers {
		go handler(data)
	}
}

// 广播设备状态更新
func BroadcastDeviceStatusUpdate(deviceID string, status string) {
	if GlobalHub != nil {
		data := map[string]interface{}{
			"device_id": deviceID,
			"status":    status,
			"timestamp": time.Now().Unix(),
		}
		GlobalHub.BroadcastMessage(MessageTypeDeviceStatusUpdate, data)
	}
}

// 广播温度数据
func BroadcastTemperatureData(data interface{}) {
	if GlobalHub != nil {
		GlobalHub.BroadcastMessage(MessageTypeTemperatureData, data)
		// 通知数据处理器
		notifyDataHandlers("temperature", data)
	}
}

// 广播断路器数据
func BroadcastBreakerData(data interface{}) {
	if GlobalHub != nil {
		GlobalHub.BroadcastMessage(MessageTypeBreakerData, data)
	}
}

// 广播服务器数据
func BroadcastServerData(data interface{}) {
	if GlobalHub != nil {
		GlobalHub.BroadcastMessage(MessageTypeServerData, data)
	}
}

// 广播告警触发
func BroadcastAlarmTriggered(data interface{}) {
	if GlobalHub != nil {
		GlobalHub.BroadcastMessage(MessageTypeAlarmTriggered, data)
	}
}

// 广播AI控制执行
func BroadcastAIControlExecuted(data interface{}) {
	if GlobalHub != nil {
		GlobalHub.BroadcastMessage(MessageTypeAIControlExecuted, data)
	}
}

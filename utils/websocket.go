package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	
	"github.com/gorilla/websocket"
)

// 定义WebSocket连接升级器
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 表示WebSocket客户端连接
type Client struct {
	ID     int64
	IP     string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// 消息类型
const (
	MessageTypeText     = "text"     // 文本消息
	MessageTypeImage    = "image"    // 图片消息
	MessageTypeEmoji    = "emoji"    // 表情消息
	MessageTypeSystem   = "system"   // 系统消息
	MessageTypeUser     = "user"     // 用户信息更新
	MessageTypeUsers    = "users"    // 在线用户列表
	MessageTypeStats    = "stats"    // 聊天室统计信息
	MessageTypeFile     = "file"     // 文件消息
	MessageTypeRecall   = "recall"   // 消息撤回
)

// Message 代表从客户端发送或接收的消息
type Message struct {
	Type      string      `json:"type"`
	Content   string      `json:"content,omitempty"`
	Username  string      `json:"username,omitempty"`
	UserID    int64       `json:"user_id,omitempty"`
	IP        string      `json:"ip,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	MessageID int64       `json:"message_id,omitempty"` // 消息ID，用于撤回
	FileName  string      `json:"file_name,omitempty"`  // 文件名
	FileSize  int64       `json:"file_size,omitempty"`  // 文件大小
	Status    int         `json:"status,omitempty"`     // 消息状态
}

// Hub 管理所有活动的客户端连接
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.Mutex
}

// NewHub 创建新的Hub实例
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run 启动WebSocket Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// BroadcastMessage 向所有客户端广播消息
func (h *Hub) BroadcastMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("错误: 消息序列化失败: %v", err)
		return
	}
	
	h.broadcast <- data
} 
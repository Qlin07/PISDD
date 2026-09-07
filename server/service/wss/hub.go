package wss

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"

	"simplechat/server/model"
)

// Client 单个WebSocket连接(一个用户可多端在线, 同一用户多个连接)
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub 维护所有在线连接
type Hub struct {
	mu       sync.RWMutex
	clients  map[int64]map[*Client]bool // userID -> 连接集合
	// 会话成员映射: conversationID -> userID 集合(发送时计算接收方)
	convUsers map[int64]map[int64]bool
}

var hub *Hub

func init() {
	hub = &Hub{
		clients:   make(map[int64]map[*Client]bool),
		convUsers: make(map[int64]map[int64]bool),
	}
}

// GetHub 获取单例
func GetHub() *Hub { return hub }

// Register 注册连接
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]bool)
	}
	h.clients[c.UserID][c] = true
	h.mu.Unlock()
}

// Unregister 注销连接
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if set, ok := h.clients[c.UserID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, c.UserID)
		}
	}
	h.mu.Unlock()
}

// RegisterConv 登记会话成员
func (h *Hub) RegisterConv(convID int64, userIDs []int64) {
	h.mu.Lock()
	h.convUsers[convID] = make(map[int64]bool)
	for _, u := range userIDs {
		h.convUsers[convID][u] = true
	}
	h.mu.Unlock()
}

// ConvUserIDs 获取会话成员ID(仅在本进程有效; MVP单实例足够)
func (h *Hub) ConvUserIDs(convID int64) []int64 {
	h.mu.RLock()
	m, ok := h.convUsers[convID]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	out := make([]int64, 0, len(m))
	for u := range m {
		out = append(out, u)
	}
	return out
}

// IsOnline 用户是否有在线连接
func (h *Hub) IsOnline(userID int64) (bool, int) {
	h.mu.RLock()
	set, ok := h.clients[userID]
	n := len(set)
	h.mu.RUnlock()
	return ok && n > 0, n
}

// SendToUser 向某用户所有连接推送消息
func (h *Hub) SendToUser(userID int64, action string, data interface{}) {
	payload, err := buildFrame(action, data)
	if err != nil {
		log.Println("ws frame error:", err)
		return
	}
	h.mu.RLock()
	set := h.clients[userID]
	conns := make([]*Client, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	for _, c := range conns {
		select {
		case c.Send <- payload:
		default:
		}
	}
}

// BroadcastToConv 推送消息到会话内用户(group中排除发送者处理由调用方决定)
func (h *Hub) BroadcastToConv(convID int64, excludeUserID int64, action string, data interface{}) {
	payload, err := buildFrame(action, data)
	if err != nil {
		log.Println("ws frame error:", err)
		return
	}
	h.mu.RLock()
	users := h.convUsers[convID]
	conns := make([]*Client, 0)
	for uid := range users {
		if uid == excludeUserID {
			continue
		}
		for c := range h.clients[uid] {
			conns = append(conns, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range conns {
		select {
		case c.Send <- payload:
		default:
		}
	}
}

// SendToMany 向多个用户推送
func (h *Hub) SendToMany(userIDs []int64, action string, data interface{}) {
	for _, uid := range userIDs {
		h.SendToUser(uid, action, data)
	}
}

func buildFrame(action string, data interface{}) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"action": action, "data": data})
}

// 消息读取写入循环
func (c *Client) ReadPump(onMessage func(c *Client, raw []byte)) {
	defer func() {
		c.Conn.Close()
		hub.Unregister(c)
	}()
	c.Conn.SetReadLimit(64 * 1024 * 1024) // 文件/图片等大消息体
	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		onMessage(c, raw)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// PendingMessage 消息服务返回给客户端的结构
type PendingMessage struct {
	*model.Message
	SenderNickname string `json:"sender_nickname"`
	SenderAvatar   string `json:"sender_avatar"`
}
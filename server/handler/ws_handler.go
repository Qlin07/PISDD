package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/service"
	"simplechat/server/service/wss"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境放行; 生产应按白名单校验Origin
	},
}

// WsHandler WebSocket长连接
type WsHandler struct {
	msg *service.MessageService
}

func NewWsHandler() *WsHandler {
	return &WsHandler{msg: service.NewMessageService()}
}

// Ws 升级为WebSocket连接(需已通过JWT鉴权)
func (h *WsHandler) Ws(c *gin.Context) {
	userID := uid(c)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &wss.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 128),
	}
	// 标记在线
	dao.DB.Model(&model.User{}).Where("user_id = ?", userID).Update("status", 1)

	wss.GetHub().Register(client)
	go client.WritePump()
	// 通知好友在线
	client.ReadPump(h.msg.HandleWS)

	// 断开置离线
	dao.DB.Model(&model.User{}).Where("user_id = ?", userID).Update("status", 0)
}
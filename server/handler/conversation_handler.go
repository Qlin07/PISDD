package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"simplechat/server/service"
)

// ConversationHandler 会话+消息
type ConversationHandler struct {
	conv    *service.ConversationService
	msg     *service.MessageService
	search  *service.SearchService
}

func NewConversationHandler() *ConversationHandler {
	return &ConversationHandler{
		conv:   service.NewConversationService(),
		msg:    service.NewMessageService(),
		search: service.NewSearchService(),
	}
}

func (h *ConversationHandler) List(c *gin.Context) {
	list, err := h.conv.List(uid(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

type singleReq struct {
	PeerID int64 `json:"peer_id"`
}

func (h *ConversationHandler) EnsureSingle(c *gin.Context) {
	var req singleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	convID, err := h.conv.EnsureSingle(uid(c), req.PeerID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	h.conv.RegisterConvWS(convID)
	OK(c, gin.H{"conversation_id": convID})
}

func (h *ConversationHandler) Top(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	topVal := c.Query("top") == "1"
	if err := h.conv.SetTop(uid(c), id, topVal); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *ConversationHandler) Mute(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	muteVal := c.Query("mute") == "1"
	if err := h.conv.SetMute(uid(c), id, muteVal); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *ConversationHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.conv.Delete(uid(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *ConversationHandler) History(c *gin.Context) {
	convID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	beforeID, _ := strconv.ParseInt(c.DefaultQuery("before_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if !h.conv.IsMember(convID, uid(c)) {
		BadRequest(c, "非会话成员")
		return
	}
	msgs, _, err := h.msg.History(uid(c), convID, beforeID, limit)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, msgs)
}

func (h *ConversationHandler) MarkRead(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.conv.MarkRead(uid(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *ConversationHandler) Search(c *gin.Context) {
	kw := c.Query("keyword")
	result, err := h.search.Search(uid(c), kw, 50)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, result)
}
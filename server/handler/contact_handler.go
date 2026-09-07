package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"simplechat/server/service"
)

// ContactHandler 联系人
type ContactHandler struct {
	contact *service.ContactService
}

func NewContactHandler() *ContactHandler {
	return &ContactHandler{contact: service.NewContactService()}
}

func (h *ContactHandler) Search(c *gin.Context) {
	kw := c.Query("keyword")
	friends, err := h.contact.Search(kw, uid(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, friends)
}

type applyReq struct {
	ToID   int64   `json:"to_id"`
	Remark *string `json:"remark"`
}

func (h *ContactHandler) Apply(c *gin.Context) {
	var req applyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.contact.SendApply(uid(c), req.ToID, req.Remark); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

type acceptReq struct {
	FromID int64  `json:"from_id"`
	Remark string `json:"remark"`
}

func (h *ContactHandler) Accept(c *gin.Context) {
	var req acceptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.contact.AcceptApply(uid(c), req.FromID, req.Remark); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

type rejectReq struct {
	FromID int64 `json:"from_id"`
}

func (h *ContactHandler) Reject(c *gin.Context) {
	var req rejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.contact.RejectApply(uid(c), req.FromID); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *ContactHandler) Pending(c *gin.Context) {
	list, err := h.contact.PendingApplies(uid(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (h *ContactHandler) Friends(c *gin.Context) {
	list, err := h.contact.ListFriendObjects(uid(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (h *ContactHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.contact.RemoveFriend(uid(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

type remarkReq struct {
	Remark string `json:"remark"`
}

func (h *ContactHandler) Remark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "参数错误")
		return
	}
	var req remarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.contact.SetRemark(uid(c), id, req.Remark); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}
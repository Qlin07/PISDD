package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"simplechat/server/service"
)

// GroupHandler 群组
type GroupHandler struct {
	group   *service.GroupService
	conv    *service.ConversationService
	contact *service.ContactService
}

func NewGroupHandler() *GroupHandler {
	return &GroupHandler{
		group:   service.NewGroupService(),
		conv:    service.NewConversationService(),
		contact: service.NewContactService(),
	}
}

type createGroupReq struct {
	Name          string  `json:"name"`
	Announcement  string  `json:"announcement"`
	MemberIDs     []int64 `json:"member_ids"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req createGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	g, err := h.group.Create(uid(c), req.Name, req.Announcement, req.MemberIDs)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, g)
}

func (h *GroupHandler) Info(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "参数错误")
		return
	}
	g, err := h.group.GetByID(id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, g)
}

// group成员列表返回(群详情带成员)
func (h *GroupHandler) Members(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "参数错误")
		return
	}
	list, err := h.group.Members(id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	userIDs := make([]int64, 0, len(list))
	for _, m := range list {
		userIDs = append(userIDs, m.UserID)
	}
	users, _ := h.contact.BatchGetUsers(userIDs)
	OK(c, gin.H{"members": list, "users": users})
}

type memberIDReq struct {
	UserID int64 `json:"user_id"`
}

func (h *GroupHandler) AddMember(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req memberIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.group.AddMember(id, uid(c), req.UserID); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req memberIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.group.RemoveMember(id, uid(c), req.UserID); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *GroupHandler) Quit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.group.Quit(id, uid(c)); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *GroupHandler) Dismiss(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.group.Dismiss(id, uid(c)); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

type updateGroupReq struct {
	Name         *string `json:"name"`
	Announcement *string `json:"announcement"`
	AvatarURL    *string `json:"avatar_url"`
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req updateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.group.UpdateProfile(id, uid(c), req.Name, req.Announcement, req.AvatarURL); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, nil)
}

func (h *GroupHandler) MyGroups(c *gin.Context) {
	list, err := h.group.MyGroups(uid(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}
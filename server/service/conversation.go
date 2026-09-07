package service

import (
	"errors"
	"time"

	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/service/wss"
	"simplechat/server/util"
)

// ConversationService 会话管理
type ConversationService struct{}

func NewConversationService() *ConversationService { return &ConversationService{} }

// EnsureSingle 获取或创建单聊会话
func (s *ConversationService) EnsureSingle(userID, peerID int64) (int64, error) {
	if userID == peerID {
		return 0, errors.New("不能和自己聊天")
	}
	// 查找已存在的单聊会话(双向, 通过共同会话成员判断)
	q := `SELECT c.conversation_id FROM conversation c
	      INNER JOIN conversation_member m1 ON m1.conversation_id = c.conversation_id AND m1.user_id = ?
	      INNER JOIN conversation_member m2 ON m2.conversation_id = c.conversation_id AND m2.user_id = ?
	      WHERE c.type = 0 LIMIT 1`
	var convID int64
	row := dao.DB.Raw(q, userID, peerID).Row()
	if err := row.Scan(&convID); err == nil && convID > 0 {
		return convID, nil
	}
	// 创建
	now := time.Now()
	conv := model.Conversation{
		ConversationID: util.GenID(),
		Type:           0,
		PeerUserID:     &peerID,
		CreatedAt:      now,
	}
	if err := dao.DB.Create(&conv).Error; err != nil {
		return 0, err
	}
	for _, uid := range []int64{userID, peerID} {
		dao.DB.Create(&model.ConversationMember{
			ConversationID: conv.ConversationID, UserID: uid, JoinedAt: now,
		})
	}
	return conv.ConversationID, nil
}

// GetConvByGroupID 获取群里会话
func (s *ConversationService) GetConvByGroupID(groupID int64) (int64, error) {
	var conv model.Conversation
	if err := dao.DB.Where("group_id = ? AND type = 1", groupID).First(&conv).Error; err != nil {
		return 0, err
	}
	return conv.ConversationID, nil
}

// IsMuted 会话是否免打扰
func (s *ConversationService) IsMuted(convID, userID int64) bool {
	var m model.ConversationMember
	if err := dao.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).First(&m).Error; err != nil {
		return false
	}
	return m.IsMute == 1
}

// IsMember 是否会话成员
func (s *ConversationService) IsMember(convID, userID int64) bool {
	var count int64
	dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).Count(&count)
	return count > 0
}

// ConvUserIDs 会话内用户ID
func (s *ConversationService) ConvUserIDs(convID int64) ([]int64, error) {
	var ms []model.ConversationMember
	if err := dao.DB.Where("conversation_id = ?", convID).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.UserID)
	}
	return out, nil
}

// List 我的会话列表(含会话信息/对方信息/未读)
type ConvVO struct {
	model.Conversation
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	UnreadCount int    `json:"unread_count"`
	IsTop       int8   `json:"is_top"`
	IsMute      int8   `json:"is_mute"`
	TypeName    string `json:"type_name"`
	PeerUserID  int64  `json:"peer_user_id,omitempty"`
}

func (s *ConversationService) List(userID int64) ([]ConvVO, error) {
	var members []model.ConversationMember
	if err := dao.DB.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, nil
	}
	convIDs := make([]int64, 0, len(members))
	convMap := map[int64]model.ConversationMember{}
	for _, m := range members {
		convIDs = append(convIDs, m.ConversationID)
		convMap[m.ConversationID] = m
	}
	var convs []model.Conversation
	if err := dao.DB.Where("conversation_id IN ?", convIDs).Find(&convs).Error; err != nil {
		return nil, err
	}
	// 收集需要展示的用户ID(单聊对端)与群
	peerIDs := []int64{}
	groupIDs := []int64{}
	for _, c := range convs {
		if c.Type == 0 && c.PeerUserID != nil {
			peerIDs = append(peerIDs, *c.PeerUserID)
		}
		if c.Type == 1 && c.GroupID != nil {
			groupIDs = append(groupIDs, *c.GroupID)
		}
	}
	userMap := map[int64]model.User{}
	if len(peerIDs) > 0 {
		var us []model.User
		dao.DB.Where("user_id IN ?", peerIDs).Find(&us)
		for _, u := range us {
			userMap[u.UserID] = u
		}
	}
	groupMap := map[int64]model.Group{}
	if len(groupIDs) > 0 {
		var gs []model.Group
		dao.DB.Where("group_id IN ?", groupIDs).Find(&gs)
		for _, g := range gs {
			groupMap[g.GroupID] = g
		}
	}
	out := make([]ConvVO, 0, len(convs))
	for _, c := range convs {
		m := convMap[c.ConversationID]
		vo := ConvVO{Conversation: c, UnreadCount: m.UnreadCount, IsTop: m.IsTop, IsMute: m.IsMute}
		if c.Type == 0 {
			vo.TypeName = "single"
			if c.PeerUserID != nil {
				if u, ok := userMap[*c.PeerUserID]; ok {
					vo.DisplayName = u.Nickname
					vo.Avatar = deref(u.AvatarURL)
					vo.PeerUserID = *c.PeerUserID
				}
			}
		} else {
			vo.TypeName = "group"
			if c.GroupID != nil {
				if g, ok := groupMap[*c.GroupID]; ok {
					vo.DisplayName = g.GroupName
					vo.Avatar = deref(g.AvatarURL)
				}
			}
		}
		out = append(out, vo)
	}
	// 排序: 置顶优先, 其次最后消息时间(desc)
	sortConvs(out)
	return out, nil
}

// SetTop 置顶/取消
func (s *ConversationService) SetTop(userID, convID int64, top bool) error {
	if !s.IsMember(convID, userID) {
		return errors.New("非会话成员")
	}
	val := 0
	if top {
		// 置顶上限10
		var count int64
		dao.DB.Model(&model.ConversationMember{}).
			Where("user_id = ? AND is_top = 1", userID).Count(&count)
		if count >= 10 {
			// 允许覆盖: 取消最旧置顶(简化处理)
		}
		val = 1
	}
	return dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("is_top", val).Error
}

// SetMute 免打扰开关
func (s *ConversationService) SetMute(userID, convID int64, mute bool) error {
	if !s.IsMember(convID, userID) {
		return errors.New("非会话成员")
	}
	val := 0
	if mute {
		val = 1
	}
	return dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("is_mute", val).Error
}

// Delete 删除会话(仅移出列表, C端+DB删除成员行)
func (s *ConversationService) Delete(userID, convID int64) error {
	return dao.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).
		Delete(&model.ConversationMember{}).Error
}

// RegisterConvWS 将会话成员登记到WS Hub(进程内)
func (s *ConversationService) RegisterConvWS(convID int64) {
	if ids, err := s.ConvUserIDs(convID); err == nil {
		wss.GetHub().RegisterConv(convID, ids)
	}
}

// MarkRead 标记会话已读并返回未读消息上一条
func (s *ConversationService) MarkRead(userID, convID int64) error {
	// 置0
	return dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("unread_count", 0).Error
}

// SortConv 供排序
func sortConvs(list []ConvVO) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0; j-- {
			a, b := list[j-1], list[j]
			// 置顶优先
			if a.IsTop != b.IsTop {
				if a.IsTop < b.IsTop {
					list[j-1], list[j] = b, a
				}
				continue
			}
			// 最后消息时间desc
			at := time.Time{}
			bt := time.Time{}
			if a.LastMsgTime != nil {
				at = *a.LastMsgTime
			}
			if b.LastMsgTime != nil {
				bt = *b.LastMsgTime
			}
			if at.Before(bt) {
				list[j-1], list[j] = b, a
			}
		}
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
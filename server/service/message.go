package service

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/service/wss"
	"simplechat/server/util"
)

var errNotMember = errors.New("非会话成员")

// WsInbound 客户端经WebSocket发来的动作
type WsInbound struct {
	Action         string `json:"action"` // send / read / typing
	ConversationID int64  `json:"conversation_id"`
	Type           int8   `json:"type"` // 0文本 1图片 2文件 3表情
	Content        string `json:"content"`
	MediaURL       string `json:"media_url"`
	MessageID      int64  `json:"message_id"` // 已读时传入
	MsgID          string `json:"client_msg_id"`
}

// MessageResult 推送/返回给客户端的消息体
type MessageResult struct {
	*model.Message
	SenderNickname string `json:"sender_nickname"`
	SenderAvatar   string `json:"sender_avatar"`
}

// MessageService 消息服务
type MessageService struct {
	conv *ConversationService
}

func NewMessageService() *MessageService {
	return &MessageService{conv: NewConversationService()}
}

// HandleWS 处理WebSocket入站消息(读取循环回调)
func (s *MessageService) HandleWS(c *wss.Client, raw []byte) {
	var in WsInbound
	if err := json.Unmarshal(raw, &in); err != nil {
		return
	}
	switch in.Action {
	case "send":
		s.HandleSend(c.UserID, in)
	case "read":
		s.HandleRead(c.UserID, in.ConversationID, in.MessageID)
	}
}

// HandleSend 发送消息
func (s *MessageService) HandleSend(userID int64, in WsInbound) error {
	// 校验会话成员
	if !s.conv.IsMember(in.ConversationID, userID) {
		return errNotMember
	}
	now := time.Now()
	content := in.Content
	msg := &model.Message{
		MessageID:      util.GenID(),
		ConversationID: in.ConversationID,
		SenderID:       userID,
		Type:           in.Type,
		Content:        &content,
		MediaURL:       nil,
		SentTime:       now,
		Status:         1, // 已发送
	}
	if in.MediaURL != "" {
		msg.MediaURL = &in.MediaURL
	}
	if err := dao.DB.Create(msg).Error; err != nil {
		return err
	}

	// 会话内其他成员
	memberIDs, err := s.conv.ConvUserIDs(in.ConversationID)
	if err != nil {
		return err
	}
	preview := content
	if len([]rune(preview)) > 50 {
		r := []rune(preview)
		preview = string(r[:50])
	}
	// 更新会话最后消息
	dao.DB.Model(&model.Conversation{}).Where("conversation_id = ?", in.ConversationID).
		Updates(map[string]interface{}{
			"last_msg_time": now, "last_msg_id": msg.MessageID, "last_msg_preview": preview,
		})

	// 逐成员: 写入状态/未读, 推送(排除发送者)
	for _, mid := range memberIDs {
		if mid == userID {
			continue
		}
		dao.DB.Create(&model.MessageStatus{MessageID: msg.MessageID, UserID: mid, Status: 0})
		// 未读+1
		updateUnread(mid, in.ConversationID)
	}
	// 推送到会话成员(含发送者自身, 用于回显已发送)
	s.pushToConv(in.ConversationID, msg)
	return nil
}

// HandleRead 已读回执: 将该会话内发送给该用户的消息标记已读
func (s *MessageService) HandleRead(userID, convID, upToMsgID int64) error {
	// 更新 conversation_member 已读位置与未读清零
	dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Updates(map[string]interface{}{"unread_count": 0})
	// 标记 message_status 已读(该用户收到且>0的消息)
	cond := dao.DB.Model(&model.MessageStatus{}).
		Where("user_id = ? AND status = 0", userID)
	if upToMsgID > 0 {
		cond = cond.Where("message_id <= ?", upToMsgID)
	}
	readTime := time.Now()
	cond.Updates(map[string]interface{}{"status": 1, "read_time": readTime})
	return nil
}

// pushToConv 推送到会话内所有成员(含发送者回显), 依DB成员计算路由
func (s *MessageService) pushToConv(convID int64, msg *model.Message) {
	payload := MessageResult{Message: msg}
	// 发送者昵称
	if u, err := (&AuthService{}).GetPublicUser(msg.SenderID); err == nil {
		payload.SenderNickname = u.Nickname
		payload.SenderAvatar = deref(u.AvatarURL)
	}
	memberIDs, err := s.conv.ConvUserIDs(convID)
	if err != nil {
		return
	}
	for _, mid := range memberIDs {
		wss.GetHub().SendToUser(mid, "message", payload)
	}
}

// History 分页拉取历史消息
func (s *MessageService) History(userID, convID int64, beforeID int64, limit int) ([]MessageResult, []model.User, error) {
	q := dao.DB.Where("conversation_id = ?", convID)
	if beforeID > 0 {
		q = q.Where("message_id < ?", beforeID)
	}
	q = q.Order("message_id DESC").Limit(limit)
	var msgs []model.Message
	if err := q.Find(&msgs).Error; err != nil {
		return nil, nil, err
	}
	// 反转为正序
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	result := make([]MessageResult, 0, len(msgs))
	senderIDSet := map[int64]bool{}
	for i := range msgs {
		senderIDSet[msgs[i].SenderID] = true
		senderNick, senderAvatar := "", ""
		if u, err := (&AuthService{}).GetPublicUser(msgs[i].SenderID); err == nil {
			senderNick = u.Nickname
			senderAvatar = deref(u.AvatarURL)
		}
		r := MessageResult{Message: &msgs[i], SenderNickname: senderNick, SenderAvatar: senderAvatar}
		result = append(result, r)
	}
	var users []model.User
	return result, users, nil
}

// updateUnread 未读+1
func updateUnread(userID, convID int64) {
	dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + 1"))
}
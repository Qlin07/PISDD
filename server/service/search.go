package service

import (
	"simplechat/server/dao"
	"simplechat/server/model"
)

// SearchHistoryReq 搜索入参
type SearchHistoryReq struct {
	UserID int64  `json:"-"`
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
}

// SearchResult 搜索结果
type SearchResult struct {
	model.Message
	ConversationName string `json:"conversation_name"`
	SenderNickname   string `json:"sender_nickname"`
}

// SearchService 消息搜索(用户所有会话范围内)
type SearchService struct{}

func NewSearchService() *SearchService { return &SearchService{} }

func (s *SearchService) Search(userID int64, keyword string, limit int) ([]SearchResult, error) {
	if keyword == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	// 我的会话ID
	var members []model.ConversationMember
	if err := dao.DB.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, nil
	}
	convIDs := make([]int64, 0, len(members))
	for _, m := range members {
		convIDs = append(convIDs, m.ConversationID)
	}
	// 匹配文本消息(类型0/3)且未撤回
	kw := "%" + keyword + "%"
	var msgs []model.Message
	if err := dao.DB.Where("conversation_id IN ? AND type IN (0,3) AND is_recalled = 0 AND content LIKE ?",
		convIDs, kw).
		Order("message_id DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}
	// 补充会话名与发送者名
	convName := map[int64]string{}
	{
		var convs []model.Conversation
		ids := make([]int64, 0, len(msgs))
		seen := map[int64]bool{}
		for _, m := range msgs {
			if !seen[m.ConversationID] {
				seen[m.ConversationID] = true
				ids = append(ids, m.ConversationID)
			}
		}
		dao.DB.Where("conversation_id IN ?", ids).Find(&convs)
		peerIDs := []int64{}
		groupIDs := []int64{}
		for _, c := range convs {
			if c.Type == 0 && c.PeerUserID != nil {
				peerIDs = append(peerIDs, *c.PeerUserID)
			} else if c.Type == 1 && c.GroupID != nil {
				groupIDs = append(groupIDs, *c.GroupID)
			}
		}
		uMap := map[int64]model.User{}
		if len(peerIDs) > 0 {
			var us []model.User
			dao.DB.Where("user_id IN ?", peerIDs).Find(&us)
			for _, u := range us {
				uMap[u.UserID] = u
			}
		}
		gMap := map[int64]model.Group{}
		if len(groupIDs) > 0 {
			var gs []model.Group
			dao.DB.Where("group_id IN ?", groupIDs).Find(&gs)
			for _, g := range gs {
				gMap[g.GroupID] = g
			}
		}
		for _, c := range convs {
			if c.Type == 0 && c.PeerUserID != nil {
				if u, ok := uMap[*c.PeerUserID]; ok {
					convName[c.ConversationID] = u.Nickname
				}
			} else if c.Type == 1 && c.GroupID != nil {
				if g, ok := gMap[*c.GroupID]; ok {
					convName[c.ConversationID] = g.GroupName
				}
			}
		}
	}
	out := make([]SearchResult, 0, len(msgs))
	for i := range msgs {
		r := SearchResult{Message: msgs[i], ConversationName: convName[msgs[i].ConversationID]}
		if u, err := (&AuthService{}).GetPublicUser(msgs[i].SenderID); err == nil {
			r.SenderNickname = u.Nickname
		}
		out = append(out, r)
	}
	return out, nil
}
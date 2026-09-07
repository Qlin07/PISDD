package service

import (
	"errors"
	"time"

	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/util"
)

// Group角色
const (
	GroupRoleMember   = 0
	GroupRoleAdmin    = 1
	GroupRoleCreator  = 2
)

// GroupService 群聊管理
type GroupService struct{}

func NewGroupService() *GroupService { return &GroupService{} }

// Create 创建群组, creator为群主
func (s *GroupService) Create(creatorID int64, name, announcement string, memberIDs []int64) (*model.Group, error) {
	if len([]rune(name)) > 30 || name == "" {
		return nil, errors.New("群名称必填且最长30字符")
	}
	if announcement != "" && len([]rune(announcement)) > 200 {
		return nil, errors.New("群公告最长200字符")
	}
	// 去重并加入群主
	ids := []int64{creatorID}
	seen := map[int64]bool{creatorID: true}
	for _, id := range memberIDs {
		if id != creatorID && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	now := time.Now()
	g := &model.Group{
		GroupID:   util.GenID(),
		GroupName: name,
		CreatorID: creatorID,
		Announcement: strPtr(announcement),
		MemberCount: len(ids),
		CreatedAt: now,
	}
	if err := dao.DB.Create(g).Error; err != nil {
		return nil, err
	}
	// 群成员
	for i, id := range ids {
		role := int8(GroupRoleMember)
		if id == creatorID {
			role = GroupRoleCreator
		}
		if err := dao.DB.Create(&model.GroupMember{
			GroupID: g.GroupID, UserID: id, Role: role, JoinedAt: now,
		}).Error; err != nil {
			return nil, err
		}
		_ = i
	}
	// 创建群聊会话并登记成员
	s.ensureGroupConversation(g.GroupID, ids)
	return g, nil
}

// ensureGroupConversation 群组会话
func (s *GroupService) ensureGroupConversation(groupID int64, memberIDs []int64) {
	var conv model.Conversation
	err := dao.DB.Where("group_id = ? AND type = 1", groupID).First(&conv).Error
	if err != nil {
		conv = model.Conversation{
			ConversationID: util.GenID(),
			Type:           1,
			GroupID:        &groupID,
			CreatedAt:      time.Now(),
		}
		dao.DB.Create(&conv)
	}
	for _, uid := range memberIDs {
		var m model.ConversationMember
		err := dao.DB.Where("conversation_id = ? AND user_id = ?", conv.ConversationID, uid).First(&m).Error
		if err != nil {
			dao.DB.Create(&model.ConversationMember{
				ConversationID: conv.ConversationID, UserID: uid,
				JoinedAt: time.Now(),
			})
		}
	}
}

// GetByID 获取群详情
func (s *GroupService) GetByID(groupID int64) (*model.Group, error) {
	var g model.Group
	if err := dao.DB.First(&g, "group_id = ?", groupID).Error; err != nil {
		return nil, errors.New("群组不存在")
	}
	return &g, nil
}

// IsMember 是否群成员
func (s *GroupService) IsMember(groupID, userID int64) bool {
	var count int64
	dao.DB.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
	return count > 0
}

// MemberRole 获取成员角色
func (s *GroupService) MemberRole(groupID, userID int64) (int8, error) {
	var m model.GroupMember
	if err := dao.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error; err != nil {
		return 0, errors.New("非群成员")
	}
	return m.Role, nil
}

// Members 群成员列表
func (s *GroupService) Members(groupID int64) ([]model.GroupMember, error) {
	var ms []model.GroupMember
	err := dao.DB.Where("group_id = ?", groupID).Order("role DESC").Find(&ms).Error
	return ms, err
}

// GetConvID 获取群会话ID
func (s *GroupService) GetConvID(groupID int64) (int64, error) {
	var conv model.Conversation
	err := dao.DB.Where("group_id = ? AND type = 1", groupID).First(&conv).Error
	if err != nil {
		return 0, errors.New("群会话不存在")
	}
	return conv.ConversationID, nil
}

// AddMember 添加群成员
func (s *GroupService) AddMember(groupID, operatorID, newUserID int64) error {
	role, err := s.MemberRole(groupID, operatorID)
	if err != nil || (role != GroupRoleCreator && role != GroupRoleAdmin) {
		return errors.New("无操作权限，仅群主/管理员可邀请")
	}
	convID, err := s.GetConvID(groupID)
	if err != nil {
		return err
	}
	// 若已存在则忽略
	var count int64
	dao.DB.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, newUserID).Count(&count)
	if count > 0 {
		return nil
	}
	if err := dao.DB.Create(&model.GroupMember{
		GroupID: groupID, UserID: newUserID, JoinedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	dao.DB.Model(&model.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, newUserID).
		FirstOrCreate(&model.ConversationMember{
			ConversationID: convID, UserID: newUserID, JoinedAt: time.Now(),
		})
	s.recountMembers(groupID)
	return nil
}

// RemoveMember 移除成员
func (s *GroupService) RemoveMember(groupID, operatorID, targetID int64) error {
	if operatorID == targetID {
		return errors.New("退出请使用退出接口")
	}
	role, err := s.MemberRole(groupID, operatorID)
	if err != nil || (role != GroupRoleCreator && role != GroupRoleAdmin) {
		return errors.New("无操作权限")
	}
	targetRole, _ := s.MemberRole(groupID, targetID)
	if targetRole == GroupRoleCreator {
		return errors.New("不能移除群主")
	}
	dao.DB.Where("group_id = ? AND user_id = ?", groupID, targetID).Delete(&model.GroupMember{})
	convID, _ := s.GetConvID(groupID)
	dao.DB.Where("conversation_id = ? AND user_id = ?", convID, targetID).Delete(&model.ConversationMember{})
	s.recountMembers(groupID)
	return nil
}

// Quit 退出群组
func (s *GroupService) Quit(groupID, userID int64) error {
	role, err := s.MemberRole(groupID, userID)
	if err != nil {
		return err
	}
	if role == GroupRoleCreator {
		return errors.New("群主不能退出，请先解散或转让")
	}
	dao.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{})
	convID, _ := s.GetConvID(groupID)
	dao.DB.Where("conversation_id = ? AND user_id = ?", convID, userID).Delete(&model.ConversationMember{})
	s.recountMembers(groupID)
	return nil
}

// Dismiss 解散群(仅群主)
func (s *GroupService) Dismiss(groupID, userID int64) error {
	role, err := s.MemberRole(groupID, userID)
	if err != nil {
		return err
	}
	if role != GroupRoleCreator {
		return errors.New("仅群主可解散群")
	}
	return dao.DB.Model(&model.Group{}).Where("group_id = ?", groupID).
		Updates(map[string]interface{}{"is_dismissed": 1}).Error
}

// UpdateProfile 更新群资料(名称/公告/头像)
func (s *GroupService) UpdateProfile(groupID, userID int64, name, announcement, avatarURL *string) error {
	role, err := s.MemberRole(groupID, userID)
	if err != nil {
		return err
	}
	if role == GroupRoleMember {
		return errors.New("仅群主/管理员可修改群资料")
	}
	updates := map[string]interface{}{}
	if name != nil {
		updates["group_name"] = *name
	}
	if announcement != nil {
		updates["announcement"] = *announcement
	}
	if avatarURL != nil {
		updates["avatar_url"] = *avatarURL
	}
	if len(updates) == 0 {
		return nil
	}
	return dao.DB.Model(&model.Group{}).Where("group_id = ?", groupID).Updates(updates).Error
}

// MyGroups 我的群组列表
func (s *GroupService) MyGroups(userID int64) ([]model.Group, error) {
	var gs []model.Group
	err := dao.DB.
		Joins("JOIN group_member gm ON gm.group_id = `group`.group_id").
		Where("gm.user_id = ? AND `group`.is_dismissed = 0", userID).
		Find(&gs).Error
	return gs, err
}

func strPtr(s string) *string { return &s }

// recountMembers 重算并更新群成员数
func (s *GroupService) recountMembers(groupID int64) {
	var count int64
	dao.DB.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Count(&count)
	dao.DB.Model(&model.Group{}).Where("group_id = ?", groupID).
		Update("member_count", count)
}
package service

import (
	"errors"
	"time"

	"simplechat/server/dao"
	"simplechat/server/model"
)

// 好友关系状态
const (
	FriendStatusApplying = 0
	FriendStatusNormal   = 1
	FriendStatusDeleted  = 2
	FriendStatusBlocked  = 3
)

// ContactService 联系人管理
type ContactService struct{}

func NewContactService() *ContactService { return &ContactService{} }

// Search 搜索用户(按昵称模糊, 排除自己)
func (s *ContactService) Search(keyword string, excludeUserID int64) ([]model.User, error) {
	var users []model.User
	q := dao.DB
	if keyword != "" {
		q = q.Where("nickname LIKE ? OR account LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := q.Where("user_id <> ?", excludeUserID).Limit(50).Find(&users).Error
	return users, err
}

// GetByUID 按数字 UID(user_id)精确查询用户, 排除自己
func (s *ContactService) GetByUID(uid, excludeUserID int64) (*model.User, error) {
	if uid == 0 {
		return nil, errors.New("UID 不合法")
	}
	if uid == excludeUserID {
		return nil, errors.New("不能添加自己为好友")
	}
	var u model.User
	if err := dao.DB.First(&u, "user_id = ?", uid).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &u, nil
}

// SendApply 发起好友申请(双方相互建立申请记录)
func (s *ContactService) SendApply(fromID, toID int64, remark *string) error {
	if fromID == toID {
		return errors.New("不能添加自己为好友")
	}
	now := time.Now().UnixMilli()
	batch := []model.Friendship{
		{FromID: fromID, ToID: toID, Remark: remark, Status: FriendStatusApplying, CreateTime: now},
		{FromID: toID, ToID: fromID, Status: FriendStatusApplying, CreateTime: now},
	}
	for i := range batch {
		var existing model.Friendship
		err := dao.DB.Where("from_id = ? AND to_id = ?", batch[i].FromID, batch[i].ToID).First(&existing).Error
		if err == nil {
			if existing.Status == FriendStatusNormal {
				return errors.New("已是好友")
			}
			if existing.Status == FriendStatusApplying {
				return errors.New("申请已发送或等待处理")
			}
			continue // 已删除/拉黑, 重新激活
		} else if err.Error() != "record not found" {
			return err
		}
		if err := dao.DB.Create(&batch[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// AcceptApply 同意好友申请(发起方接受对方)
func (s *ContactService) AcceptApply(accepterID, requesterID int64, remark string) error {
	// 双向将申请中 -> 正常
	res := dao.DB.Model(&model.Friendship{}).
		Where("from_id = ? AND to_id = ? AND status = ?", requesterID, accepterID, FriendStatusApplying).
		Updates(map[string]interface{}{"status": FriendStatusNormal, "remark": remark})
	if res.RowsAffected == 0 {
		return errors.New("好友申请不存在或已处理")
	}
	dao.DB.Model(&model.Friendship{}).
		Where("from_id = ? AND to_id = ? AND status = ?", accepterID, requesterID, FriendStatusApplying).
		Updates(map[string]interface{}{"status": FriendStatusNormal, "remark": remark})
	return nil
}

// RejectApply 拒绝好友申请
func (s *ContactService) RejectApply(accepterID, requesterID int64) error {
	res := dao.DB.Where("from_id = ? AND to_id = ?", requesterID, accepterID).
		Delete(&model.Friendship{})
	if res.RowsAffected == 0 {
		return errors.New("申请不存在或已处理")
	}
	dao.DB.Where("from_id = ? AND to_id = ?", accepterID, requesterID).
		Delete(&model.Friendship{})
	return nil
}

// PendingApplies 我的好友申请列表
func (s *ContactService) PendingApplies(userID int64) ([]model.Friendship, error) {
	var fs []model.Friendship
	err := dao.DB.Where("to_id = ? AND status = ?", userID, FriendStatusApplying).
		Order("id DESC").Find(&fs).Error
	return fs, err
}

// ListFriends 我的好友列表(状态正常)
func (s *ContactService) ListFriends(userID int64) ([]model.Friendship, error) {
	var fs []model.Friendship
	err := dao.DB.Where("from_id = ? AND status = ?", userID, FriendStatusNormal).
		Order("id ASC").Find(&fs).Error
	return fs, err
}

// RemoveFriend 删除好友(仅移除from一方关系, 不影响对方列表——双向记录模式)
func (s *ContactService) RemoveFriend(userID, friendID int64) error {
	res := dao.DB.Where("from_id = ? AND to_id = ? AND status = ?",
		userID, friendID, FriendStatusNormal).
		Updates(map[string]interface{}{"status": FriendStatusDeleted})
	if res.RowsAffected == 0 {
		return errors.New("好友关系不存在")
	}
	return nil
}

// SetRemark 设置好友备注
func (s *ContactService) SetRemark(userID, friendID int64, remark string) error {
	return dao.DB.Model(&model.Friendship{}).
		Where("from_id = ? AND to_id = ?", userID, friendID).
		Update("remark", remark).Error
}

// BatchGetUsers 批量获取用户(供好友头像昵称展示)
func (s *ContactService) BatchGetUsers(ids []int64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	err := dao.DB.Where("user_id IN ?", ids).Find(&users).Error
	return users, err
}

// FriendVO 好友视图(用户信息+备注)
type FriendVO struct {
	model.User
	Remark string `json:"remark"`
}

// ListFriendObjects 我的好友(含备注名), 用于客户端完整展示
func (s *ContactService) ListFriendObjects(userID int64) ([]FriendVO, error) {
	var fs []model.Friendship
	if err := dao.DB.Where("from_id = ? AND status = ?", userID, FriendStatusNormal).
		Order("id ASC").Find(&fs).Error; err != nil {
		return nil, err
	}
	if len(fs) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(fs))
	for _, f := range fs {
		ids = append(ids, f.ToID)
	}
	users, err := s.BatchGetUsers(ids)
	if err != nil {
		return nil, err
	}
	remarkMap := map[int64]string{}
	for _, f := range fs {
		if f.Remark != nil {
			remarkMap[f.ToID] = *f.Remark
		}
	}
	out := make([]FriendVO, 0, len(users))
	for _, u := range users {
		out = append(out, FriendVO{User: u, Remark: remarkMap[u.UserID]})
	}
	return out, nil
}
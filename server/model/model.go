package model

import "time"

// User 用户(对应 user 表)
type User struct {
	UserID       int64      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Account      string     `gorm:"uniqueIndex;column:account" json:"account"`
	Phone        *string    `gorm:"uniqueIndex;column:phone" json:"phone"`
	Email        *string    `gorm:"uniqueIndex;column:email" json:"email"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	Nickname     string     `gorm:"column:nickname" json:"nickname"`
	AvatarURL    *string    `gorm:"column:avatar_url" json:"avatar_url"`
	Signature    *string    `gorm:"column:signature" json:"signature"`
	Status       int8       `gorm:"column:status" json:"status"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	LastActive   *time.Time `gorm:"column:last_active" json:"last_active"`
	IsVerified   int8       `gorm:"column:is_verified" json:"is_verified"`
	FailCount    int        `gorm:"column:fail_count" json:"-"`
	LockUntil    *time.Time `gorm:"column:lock_until" json:"-"`
}

func (User) TableName() string { return "user" }

// Friendship 好友关系(friendship 表)
type Friendship struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	FromID     int64  `gorm:"column:from_id" json:"from_id"`
	ToID       int64  `gorm:"column:to_id" json:"to_id"`
	Remark     *string `gorm:"column:remark" json:"remark"`
	Status     int8   `gorm:"column:status" json:"status"` // 0申请中 1正常 2已删除 3拉黑
	Black      int8   `gorm:"column:black" json:"black"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
}

func (Friendship) TableName() string { return "friendship" }

// Group 群组(group 表)
type Group struct {
	GroupID     int64     `gorm:"primaryKey;column:group_id" json:"group_id"`
	GroupName   string    `gorm:"column:group_name" json:"group_name"`
	CreatorID   int64     `gorm:"column:creator_id" json:"creator_id"`
	AvatarURL   *string   `gorm:"column:avatar_url" json:"avatar_url"`
	Announcement *string  `gorm:"column:announcement" json:"announcement"`
	MemberCount int       `gorm:"column:member_count" json:"member_count"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	IsDismissed int8      `gorm:"column:is_dismissed" json:"is_dismissed"`
}

func (Group) TableName() string { return "group" }

// GroupMember 群组成员(group_member 表)
type GroupMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID  int64     `gorm:"column:group_id" json:"group_id"`
	UserID   int64     `gorm:"column:user_id" json:"user_id"`
	Role     int8      `gorm:"column:role" json:"role"` // 0普通 1管理员 2群主
	JoinedAt time.Time `gorm:"column:joined_at" json:"joined_at"`
}

func (GroupMember) TableName() string { return "group_member" }

// Conversation 会话(conversation 表)
type Conversation struct {
	ConversationID int64      `gorm:"primaryKey;column:conversation_id" json:"conversation_id"`
	Type           int8       `gorm:"column:type" json:"type"` // 0单聊 1群聊
	PeerUserID     *int64     `gorm:"column:peer_user_id" json:"peer_user_id"`
	GroupID        *int64     `gorm:"column:group_id" json:"group_id"`
	LastMsgTime    *time.Time `gorm:"column:last_msg_time" json:"last_msg_time"`
	LastMsgID      *int64     `gorm:"column:last_msg_id" json:"last_msg_id"`
	LastMsgPreview *string    `gorm:"column:last_msg_preview" json:"last_msg_preview"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"created_at"`
}

func (Conversation) TableName() string { return "conversation" }

// ConversationMember 会话成员(conversation_member 表)
type ConversationMember struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID int64     `gorm:"column:conversation_id" json:"conversation_id"`
	UserID         int64     `gorm:"column:user_id" json:"user_id"`
	UnreadCount    int       `gorm:"column:unread_count" json:"unread_count"`
	IsTop          int8      `gorm:"column:is_top" json:"is_top"`
	IsMute         int8      `gorm:"column:is_mute" json:"is_mute"`
	JoinedAt       time.Time `gorm:"column:joined_at" json:"joined_at"`
	LastReadMsgID  *int64    `gorm:"column:last_read_msg_id" json:"last_read_msg_id"`
}

func (ConversationMember) TableName() string { return "conversation_member" }

// Message 消息(message 表)
type Message struct {
	MessageID      int64     `gorm:"primaryKey;column:message_id" json:"message_id"`
	ConversationID int64     `gorm:"column:conversation_id" json:"conversation_id"`
	SenderID       int64     `gorm:"column:sender_id" json:"sender_id"`
	Type           int8      `gorm:"column:type" json:"type"` // 0文本 1图片 2文件 3表情 4系统
	Content        *string   `gorm:"column:content" json:"content"`
	MediaURL       *string   `gorm:"column:media_url" json:"media_url"`
	FileID         *int64    `gorm:"column:file_id" json:"file_id"`
	SentTime       time.Time `gorm:"column:sent_time" json:"sent_time"`
	Status         int8      `gorm:"column:status" json:"status"` // 0发送中 1已发送 2已送达 3已读
	IsRecalled     int8      `gorm:"column:is_recalled" json:"is_recalled"`
}

func (Message) TableName() string { return "message" }

// MessageStatus 消息状态(message_status 表)
type MessageStatus struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID int64      `gorm:"column:message_id" json:"message_id"`
	UserID    int64      `gorm:"column:user_id" json:"user_id"`
	Status    int8       `gorm:"column:status" json:"status"` // 0未读 1已读
	ReadTime  *time.Time `gorm:"column:read_time" json:"read_time"`
}

func (MessageStatus) TableName() string { return "message_status" }

// OfflineMessage 离线消息(offline_message 表)
type OfflineMessage struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"user_id"`
	MessageID  int64     `gorm:"column:message_id" json:"message_id"`
	DelFlag    int8      `gorm:"column:del_flag" json:"del_flag"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
}

func (OfflineMessage) TableName() string { return "offline_message" }

// FileRecord 文件记录(file_record 表)
type FileRecord struct {
	FileID     int64     `gorm:"primaryKey;column:file_id" json:"file_id"`
	FileName   string    `gorm:"column:file_name" json:"file_name"`
	FileSize   int64     `gorm:"column:file_size" json:"file_size"`
	FileURL    string    `gorm:"column:file_url" json:"file_url"`
	FileMD5    *string   `gorm:"column:file_md5" json:"file_md5"`
	MimeType   *string   `gorm:"column:mime_type" json:"mime_type"`
	UploaderID int64     `gorm:"column:uploader_id" json:"uploader_id"`
	UploadTime time.Time `gorm:"column:upload_time" json:"upload_time"`
}

func (FileRecord) TableName() string { return "file_record" }

// 会话内用户ID辅助类型(无表)
type ConversationPeer struct {
	ConversationID int64
	UserIDs        []int64
}
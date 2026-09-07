package service

import (
	"errors"
	"regexp"
	"time"

	"gorm.io/gorm"

	"simplechat/server/config"
	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/util"
)

var (
	ErrAccountExists = errors.New("账号已存在")
	ErrUserNotFound  = errors.New("用户不存在")
	ErrWrongPassword = errors.New("账号或密码错误")
	ErrLocked        = errors.New("账号已锁定，请稍后再试")
)

var accountRe = regexp.MustCompile(`^[A-Za-z0-9]{3,32}$`)

// AuthService 用户认证
type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService { return &AuthService{cfg: cfg} }

// Register 注册
func (s *AuthService) Register(account, nickname, password string) (*model.User, error) {
	if !accountRe.MatchString(account) {
		return nil, errors.New("账号仅支持3-32位大小写字母和数字")
	}
	if nickname == "" {
		nickname = account
	}
	if len([]rune(nickname)) > 20 {
		return nil, errors.New("昵称最长20字符")
	}
	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		UserID:       util.GenID(),
		Account:      account,
		PasswordHash: hash,
		Nickname:     nickname,
		Status:       0,
		CreatedAt:    time.Now(),
	}
	err = dao.DB.Create(u).Error
	if err != nil {
		// 唯一键冲突
		return nil, ErrAccountExists
	}
	return u, nil
}

// Login 登录, remember=true 延长令牌30天
func (s *AuthService) Login(account, password string, remember bool) (string, *model.User, error) {
	var u model.User
	err := dao.DB.Where("account = ?", account).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil, ErrUserNotFound
	}
	if err != nil {
		return "", nil, err
	}
	// 锁定检查
	if u.LockUntil != nil && time.Now().Before(*u.LockUntil) {
		return "", nil, ErrLocked
	}
	if !util.CheckPassword(u.PasswordHash, password) {
		// 连续失败锁定
		newCount := u.FailCount + 1
		lockUntil := (*time.Time)(nil)
		if newCount >= 5 {
			t := time.Now().Add(15 * time.Minute)
			lockUntil = &t
			newCount = 0
		}
		dao.DB.Model(&u).Updates(map[string]interface{}{
			"fail_count": newCount, "lock_until": lockUntil,
		})
		return "", nil, ErrWrongPassword
	}
	// 成功清零
	dao.DB.Model(&u).Updates(map[string]interface{}{
		"fail_count": 0, "lock_until": nil, "status": 1, "last_active": time.Now(),
	})
	days := s.cfg.Server.JWTExpireDays
	if remember {
		days = 30
	}
	token, err := util.MakeToken(s.cfg.Server.JWTSecret, u.UserID, u.Account, days)
	if err != nil {
		return "", nil, err
	}
	return token, &u, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID int64, oldPw, newPw string) error {
	var u model.User
	if err := dao.DB.First(&u, "user_id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}
	if !util.CheckPassword(u.PasswordHash, oldPw) {
		return errors.New("原密码错误")
	}
	hash, err := util.HashPassword(newPw)
	if err != nil {
		return err
	}
	return dao.DB.Model(&u).Update("password_hash", hash).Error
}

// GetByID 获取用户信息
func (s *AuthService) GetPublicUser(userID int64) (*model.User, error) {
	var u model.User
	if err := dao.DB.First(&u, "user_id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

// UpdateProfile 编辑资料(昵称/头像/签名)
func (s *AuthService) UpdateProfile(userID int64, nickname, avatarURL, signature string) error {
	if len([]rune(nickname)) > 20 {
		return errors.New("昵称最长20字符")
	}
	if len([]rune(signature)) > 50 {
		return errors.New("个性签名最长50字符")
	}
	return dao.DB.Model(&model.User{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"nickname":   nickname,
			"avatar_url": avatarURL,
			"signature":  signature,
		}).Error
}

// Logout 登出
func (s *AuthService) Logout(userID int64) error {
	return dao.DB.Model(&model.User{}).Where("user_id = ?", userID).Update("status", 0).Error
}
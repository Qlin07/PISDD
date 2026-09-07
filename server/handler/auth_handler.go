package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"simplechat/server/config"
	"simplechat/server/service"
)

// uid 从context读取当前登录用户ID
func uid(c *gin.Context) int64 {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

func srvErrf(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrAccountExists):
		BadRequest(c, "账号已存在")
	case errors.Is(err, service.ErrUserNotFound):
		BadRequest(c, "用户不存在")
	case errors.Is(err, service.ErrWrongPassword):
		BadRequest(c, "账号或密码错误")
	case errors.Is(err, service.ErrLocked):
		BadRequest(c, "账号已锁定，请稍后再试")
	default:
		BadRequest(c, err.Error())
	}
}

// AuthHandler 认证
type AuthHandler struct {
	auth *service.AuthService
	cfg  *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: service.NewAuthService(cfg), cfg: cfg}
}

type registerReq struct {
	Account  string `json:"account"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	u, err := h.auth.Register(req.Account, req.Nickname, req.Password)
	if err != nil {
		srvErrf(c, err)
		return
	}
	// 注册即签发token并自动登录
	token, _, err := h.auth.Login(req.Account, req.Password, false)
	if err != nil {
		ServerError(c, "自动登录失败")
		return
	}
	OK(c, gin.H{"user": u, "token": token})
}

type loginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	token, u, err := h.auth.Login(req.Account, req.Password, req.Remember)
	if err != nil {
		srvErrf(c, err)
		return
	}
	OK(c, gin.H{"user": u, "token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	_ = h.auth.Logout(uid(c))
	OK(c, nil)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	u, err := h.auth.GetPublicUser(uid(c))
	if err != nil {
		srvErrf(c, err)
		return
	}
	OK(c, u)
}

type updateProfileReq struct {
	Nickname string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Signature string `json:"signature"`
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.auth.UpdateProfile(uid(c), req.Nickname, req.AvatarURL, req.Signature); err != nil {
		srvErrf(c, err)
		return
	}
	OK(c, nil)
}

type changePwdReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := h.auth.ChangePassword(uid(c), req.OldPassword, req.NewPassword); err != nil {
		srvErrf(c, err)
		return
	}
	OK(c, nil)
}
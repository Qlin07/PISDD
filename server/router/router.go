package router

import (
	"github.com/gin-gonic/gin"

	"simplechat/server/config"
	"simplechat/server/handler"
)

// Setup 路由注册
func Setup(cfg *config.Config) *gin.Engine {
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), handler.CORS(cfg.Server.WebOrigin))

	auth := handler.NewAuthHandler(cfg)
	contact := handler.NewContactHandler()
	group := handler.NewGroupHandler()
	conv := handler.NewConversationHandler()
	file := handler.NewFileHandler(cfg)
	ws := handler.NewWsHandler()

	// 无需鉴权
	pub := r.Group("/api")
	{
		pub.POST("/auth/register", auth.Register)
		pub.POST("/auth/login", auth.Login)
	}

	// 需鉴权
	api := r.Group("/api", handler.AuthMiddleware(cfg.Server.JWTSecret))
	{
		// 认证
		api.POST("/auth/logout", auth.Logout)
		api.GET("/user/profile", auth.Profile)
		api.PUT("/user/profile", auth.UpdateProfile)
		api.PUT("/user/password", auth.ChangePassword)

		// 联系人
		api.GET("/contacts/search", contact.Search)
		api.GET("/contacts/uid", contact.GetByUID)
		api.POST("/contacts/apply", contact.Apply)
		api.POST("/contacts/accept", contact.Accept)
		api.POST("/contacts/reject", contact.Reject)
		api.GET("/contacts/pending", contact.Pending)
		api.GET("/contacts", contact.Friends)
		api.DELETE("/contacts/:id", contact.Remove)
		api.PUT("/contacts/:id/remark", contact.Remark)

		// 群组
		api.POST("/groups", group.Create)
		api.GET("/groups/mine", group.MyGroups)
		api.GET("/groups/:id", group.Info)
		api.GET("/groups/:id/members", group.Members)
		api.POST("/groups/:id/members", group.AddMember)
		api.DELETE("/groups/:id/members", group.RemoveMember)
		api.POST("/groups/:id/quit", group.Quit)
		api.POST("/groups/:id/dismiss", group.Dismiss)
		api.PUT("/groups/:id", group.Update)

		// 会话 + 消息 + 搜索
		api.GET("/conversations", conv.List)
		api.POST("/conversations/single", conv.EnsureSingle)
		api.PUT("/conversations/:id/top", conv.Top)
		api.PUT("/conversations/:id/mute", conv.Mute)
		api.DELETE("/conversations/:id", conv.Delete)
		api.GET("/conversations/:id/messages", conv.History)
		api.POST("/conversations/:id/read", conv.MarkRead)
		api.GET("/search/messages", conv.Search)

		// 文件
		api.POST("/files/upload", file.Upload)
	}

	// WebSocket(走鉴权后的 token)
	wsg := r.Group("/ws", handler.AuthMiddleware(cfg.Server.JWTSecret))
	{
		wsg.GET("", ws.Ws)
	}

	return r
}
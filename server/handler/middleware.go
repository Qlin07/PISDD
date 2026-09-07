package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"simplechat/server/util"
)

// CtxUserID middleware 存入的user id key
const CtxUserID = "uid"

// AuthMiddleware JWT鉴权中间件(支持 Header 或 query token, 后者用于WebSocket握手)
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := ""
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		}
		if tokenStr == "" {
			tokenStr = c.Query("token") // WebSocket 握手通过query传token
		}
		if tokenStr == "" {
			Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		claims, err := util.ParseToken(secret, tokenStr)
		if err != nil {
			Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}
}

// CORS 跨域中间件
func CORS(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		o := c.GetHeader("Origin")
		if origin == "*" || o == origin {
			c.Header("Access-Control-Allow-Origin", o)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
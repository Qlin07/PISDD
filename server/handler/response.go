package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OK 返回成功数据
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

// Fail 返回错误
func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, gin.H{"code": code, "msg": msg, "data": nil})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}

// ServerError 服务端错误
func ServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 500, msg)
}
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess      = 0  // 业务成功，前端判定 code === 0
	CodeFail         = 1  // 一般业务失败
	CodeUnauthorized = -1 // token 无效或未登录
)

// Body 统一响应体。前端成功取 data；提示文案同时给 message 和 msg，兼容 Vben。
type Body struct {
	Code    int         `json:"code"`    // 业务码，0 成功
	Data    interface{} `json:"data"`    // 业务数据
	Message string      `json:"message"` // 提示信息，前端优先读这个
	Msg     string      `json:"msg"`     // 旧字段，内容和 message 相同
}

// Success 返回成功，提示为 ok。
func Success(c *gin.Context, data interface{}) {
	write(c, http.StatusOK, CodeSuccess, data, "ok")
}

// SuccessMsg 返回成功，并自定义提示文案。
func SuccessMsg(c *gin.Context, data interface{}, msg string) {
	write(c, http.StatusOK, CodeSuccess, data, msg)
}

// Fail 返回业务失败。HTTP 仍是 200，由 code != 0 表示失败。
func Fail(c *gin.Context, msg string) {
	write(c, http.StatusOK, CodeFail, nil, msg)
}

// FailWithCode 返回业务失败，并指定业务码和 data。
func FailWithCode(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:    code,
		Data:    data,
		Message: msg,
		Msg:     msg,
	})
}

// Unauthorized 返回未登录或 token 失效。HTTP 401，前端会跳回登录页。
func Unauthorized(c *gin.Context, msg string) {
	write(c, http.StatusUnauthorized, CodeUnauthorized, nil, msg)
}

// write 实际写出 JSON 响应。
func write(c *gin.Context, httpStatus, code int, data interface{}, msg string) {
	c.JSON(httpStatus, Body{
		Code:    code,
		Data:    data,
		Message: msg,
		Msg:     msg,
	})
}

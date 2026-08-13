package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess = 0
	CodeFail    = 1
)

type Body struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: CodeSuccess,
		Data: data,
		Msg:  "success",
	})
}

// 成功响应带消息
func SuccessMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Body{
		Code: CodeSuccess,
		Data: data,
		Msg:  msg,
	})
}

// 失败响应
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Body{
		Code: CodeFail,
		Data: nil,
		Msg:  msg,
	})
}

// 失败响应带代码和数据
func FailWithCode(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

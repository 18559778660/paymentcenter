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

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: CodeSuccess,
		Data: data,
		Msg:  "success",
	})
}

func SuccessMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Body{
		Code: CodeSuccess,
		Data: data,
		Msg:  msg,
	})
}

func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Body{
		Code: CodeFail,
		Data: nil,
		Msg:  msg,
	})
}

func FailWithCode(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

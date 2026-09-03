package response

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ShopyyErrorCodeSuccess = 0 //成功
	ShopyyErrorCodeFail    = 1 //失败

	ShopyyDataCodeFail    = 1 //失败
	ShopyyDataCodeSuccess = 2 //成功

	ShopyyPayHTMLType = 1 //HTML类型
	ShopyyVersion     = 2 //版本
)

// ShopyyPayData Shopyy 支付成功 data 节点。
type ShopyyPayData struct {
	Code        int64  `json:"code"`
	PayCode     int64  `json:"pay_code"`
	PayHTMLType int64  `json:"pay_html_type"`
	PaySkipURL  string `json:"pay_skip_url"`
	Version     int64  `json:"version"`
}

// ShopyyPayBody Shopyy 支付接口响应体。
type ShopyyPayBody struct {
	ErrorCode int64          `json:"error_code"`
	Data      *ShopyyPayData `json:"data"`
	Reason    string         `json:"reason"`
}

// ShopyyPaySuccess 返回 Shopyy 支付成功响应。
func ShopyyPaySuccess(c *gin.Context, checkoutURL string, siteMode string) {
	writeShopyyPay(c, ShopyyErrorCodeSuccess, &ShopyyPayData{
		Code:        ShopyyDataCodeSuccess,
		PayCode:     shopyySiteModeInt(siteMode),
		PayHTMLType: ShopyyPayHTMLType,
		PaySkipURL:  checkoutURL,
		Version:     ShopyyVersion,
	}, "Success")
}

// ShopyyPayFail 返回 Shopyy 支付失败响应。
func ShopyyPayFail(c *gin.Context, reason string) {
	if strings.TrimSpace(reason) == "" {
		reason = "Failed"
	}
	writeShopyyPay(c, ShopyyErrorCodeFail, nil, reason)
}

// shopyySiteModeInt 转换 Shopyy 站点模式为整数。
func shopyySiteModeInt(siteMode string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(siteMode), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// writeShopyyPay 写入 Shopyy 支付响应。
func writeShopyyPay(c *gin.Context, errorCode int64, data *ShopyyPayData, reason string) {
	c.JSON(http.StatusOK, ShopyyPayBody{
		ErrorCode: errorCode,
		Data:      data,
		Reason:    reason,
	})
}

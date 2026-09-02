package response

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ShopyyErrorCodeSuccess = 0
	ShopyyErrorCodeFail    = 1
)

// ShopyyPayData Shopyy 支付成功 data 节点。
type ShopyyPayData struct {
	URL      string `json:"url"`
	SiteMode string `json:"site_mode"`
}

// ShopyyPayBody Shopyy 支付接口响应体。
type ShopyyPayBody struct {
	ErrorCode int          `json:"error_code"`
	Data      *ShopyyPayData `json:"data"`
	Reason    string       `json:"reason"`
}

// ShopyyPaySuccess 返回 Shopyy 支付成功响应。
func ShopyyPaySuccess(c *gin.Context, checkoutURL string, siteMode string) {
	writeShopyyPay(c, ShopyyErrorCodeSuccess, &ShopyyPayData{
		URL:      checkoutURL,
		SiteMode: siteMode,
	}, "Success")
}

// ShopyyPayFail 返回 Shopyy 支付失败响应。
func ShopyyPayFail(c *gin.Context, reason string) {
	if strings.TrimSpace(reason) == "" {
		reason = "Failed"
	}
	writeShopyyPay(c, ShopyyErrorCodeFail, nil, reason)
}

func writeShopyyPay(c *gin.Context, errorCode int, data *ShopyyPayData, reason string) {
	c.JSON(http.StatusOK, ShopyyPayBody{
		ErrorCode: errorCode,
		Data:      data,
		Reason:    reason,
	})
}

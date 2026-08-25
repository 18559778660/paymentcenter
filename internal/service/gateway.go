package service

import (
	"net/url"
	"strings"
)

// GatewayAccess A 站网关探活。当前仅返回通道标识与就绪状态。
func (a *App) GatewayAccess(channel string) map[string]interface{} {
	channel = strings.TrimSpace(channel)
	return map[string]interface{}{
		"channel": channel,
		"status":  "ready",
	}
}

// BuildGatewayURL 生成 A 站对接网关地址。
func (a *App) BuildGatewayURL(channelName string) string {
	channelName = strings.TrimSpace(channelName)
	if channelName == "" || a.gatewayBaseURL == "" {
		return ""
	}
	return a.gatewayBaseURL + "/api/gateway?channel=" + url.QueryEscape(channelName)
}

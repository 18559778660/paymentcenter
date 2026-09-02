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

func (a *App) buildGatewayPayURL(query string) string {
	if a.gatewayBaseURL == "" || strings.TrimSpace(query) == "" {
		return ""
	}
	return a.gatewayBaseURL + "/api/gateway/pay?" + query
}

// BuildGatewayURL 生成 A 站发起支付地址（POST）。
func (a *App) BuildGatewayURL(channelName string) string {
	channelName = strings.TrimSpace(channelName)
	if channelName == "" {
		return ""
	}
	return a.buildGatewayPayURL("channel=" + url.QueryEscape(channelName))
}

// BuildGroupGatewayURL 生成分组发起支付地址（POST）。
func (a *App) BuildGroupGatewayURL(groupCode string) string {
	groupCode = strings.TrimSpace(groupCode)
	if groupCode == "" {
		return ""
	}
	return a.buildGatewayPayURL("group=" + url.QueryEscape(groupCode))
}

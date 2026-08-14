package controller

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/util/response"
)

// HealthController 控制层：服务健康检查。
type HealthController struct {
	cfg config.Config
}

// NewHealthController 创建健康检查控制器。
func NewHealthController(cfg config.Config) *HealthController {
	return &HealthController{cfg: cfg}
}

// Health 返回当前服务名和监听地址，用于确认服务已启动。
func (h *HealthController) Health(c *gin.Context) {
	response.Success(c, gin.H{"service": h.cfg.PaymentName, "addr": h.cfg.Addr})
}

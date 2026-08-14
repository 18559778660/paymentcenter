package controller

import (
	"github.com/gin-gonic/gin"

	"paymentcenter/internal/config"
	"paymentcenter/internal/util/response"
)

// 健康检查控制器
type HealthController struct {
	cfg config.Config
}

// 创建健康检查控制器
func NewHealthController(cfg config.Config) *HealthController {
	return &HealthController{cfg: cfg}
}

// 健康检查
func (h *HealthController) Health(c *gin.Context) {
	response.Success(c, gin.H{"service": h.cfg.PaymentName, "addr": h.cfg.Addr})
}

package main

import (
	"log"

	"github.com/joho/godotenv"

	"paymentcenter/internal/config"
	"paymentcenter/internal/service"
	"paymentcenter/internal/store"
	httptransport "paymentcenter/internal/transport/http"
)

// 支付中心入口
// 加载配置
// 创建 MySQL 存储
// 创建应用
// 创建路由
// 启动服务器
func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	st, err := store.NewMySQLStore(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}
	defer st.Close()

	app := service.NewApp(st, cfg.AuthSecret, cfg.TokenTTL, cfg.GatewayBaseURL, cfg.StripeAPIKey, cfg.StripeWebhook)
	if err := app.EnsureDefaultAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("seed admin user failed: %v", err)
	}

	router := httptransport.NewRouter(cfg, app)

	log.Printf("payment center listening on %s", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

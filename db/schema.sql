CREATE DATABASE IF NOT EXISTS payment_center
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE payment_center;

CREATE TABLE IF NOT EXISTS payment_orders (
  id VARCHAR(64) NOT NULL COMMENT '支付中心订单号',
  merchant_order VARCHAR(128) NOT NULL COMMENT 'A站订单号',
  merchant_site VARCHAR(255) NOT NULL COMMENT 'A站域名或站点标识',
  channel VARCHAR(64) NOT NULL COMMENT '通道代码，例如 win_stripe',
  provider VARCHAR(64) NOT NULL COMMENT '支付平台，例如 stripe',
  amount BIGINT NOT NULL COMMENT '最小货币单位金额，例如分/cent',
  currency VARCHAR(16) NOT NULL COMMENT '币种',
  return_url TEXT NOT NULL COMMENT '同步返回地址',
  notify_url TEXT NOT NULL COMMENT '异步通知地址',
  checkout_url TEXT NULL COMMENT '支付跳转地址',
  provider_ref VARCHAR(128) NOT NULL DEFAULT '' COMMENT '支付平台交易号',
  status VARCHAR(32) NOT NULL COMMENT 'created/pending/paid/failed/cancelled',
  error_message TEXT NULL COMMENT '失败原因',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  KEY idx_merchant_order (merchant_order),
  KEY idx_status (status),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 项目启动时也会通过 GORM AutoMigrate 自动校验/迁移表结构。

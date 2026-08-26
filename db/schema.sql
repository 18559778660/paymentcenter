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

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '后台用户ID',
  username VARCHAR(64) NOT NULL COMMENT '登录账号',
  password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt密码哈希',
  real_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '显示名',
  avatar VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像',
  home_path VARCHAR(128) NOT NULL DEFAULT '/dashboard/analytics' COMMENT '登录后首页',
  status BIGINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  last_login_at DATETIME NULL COMMENT '最后登录时间',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_users_username (username),
  KEY idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  code VARCHAR(64) NOT NULL COMMENT '角色编码，给前端用，例如 super/admin',
  name VARCHAR(64) NOT NULL COMMENT '角色名称',
  remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  status BIGINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_roles_code (code),
  KEY idx_roles_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS menus (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID，0为顶级',
  name VARCHAR(64) NOT NULL COMMENT '路由Name，需唯一',
  title VARCHAR(64) NOT NULL COMMENT '显示标题',
  path VARCHAR(128) NOT NULL DEFAULT '' COMMENT '路由路径',
  component VARCHAR(255) NOT NULL DEFAULT '' COMMENT '前端组件路径，后续动态菜单用',
  icon VARCHAR(64) NOT NULL DEFAULT '' COMMENT '图标',
  auth_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '权限码，给 /auth/codes',
  type BIGINT NOT NULL DEFAULT 1 COMMENT '0目录 1菜单 2按钮',
  sort BIGINT NOT NULL DEFAULT 0 COMMENT '排序',
  status BIGINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  meta JSON NULL COMMENT 'Vben meta 扩展字段',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_menus_name (name),
  KEY idx_menus_parent_id (parent_id),
  KEY idx_menus_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  PRIMARY KEY (user_id, role_id),
  KEY idx_user_roles_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_menus (
  role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  menu_id BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (role_id, menu_id),
  KEY idx_role_menus_menu_id (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '通道ID',
  name VARCHAR(64) NOT NULL COMMENT '通道名称',
  package_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '压缩包文件名',
  daily_order_limit INT NOT NULL DEFAULT 0 COMMENT '日限单数',
  daily_amount_limit DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '日限金额USD',
  intercept_mode VARCHAR(16) NOT NULL DEFAULT 'reset' COMMENT '拦截模式 reset重置 keep保持',
  intercept_currency VARCHAR(16) NOT NULL DEFAULT 'USD' COMMENT '拦截货币',
  intercept_max DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '拦截最大金额',
  intercept_min DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '拦截最小金额',
  status INT NOT NULL DEFAULT 1 COMMENT '状态 1启用 0禁用',
  payment_mode VARCHAR(32) NOT NULL DEFAULT 'local' COMMENT '支付模式 local本地支付 checkout收银台 embedded系统内嵌',
  remark VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
  settle_rate INT NOT NULL DEFAULT 0 COMMENT '结算比例',
  site_b_group VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'B站分组',
  channel_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通道CODE',
  pay_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '支付CODE',
  order_no_mode VARCHAR(32) NOT NULL DEFAULT 'site' COMMENT '订单号设置 site网站订单号 strip_suffix去后缀订单号',
  product_info VARCHAR(32) NOT NULL DEFAULT 'kezhan' COMMENT '商品信息 kezhan壳站信息 product产品数据 rebuild重组数据',
  return_verify VARCHAR(16) NOT NULL DEFAULT 'verify' COMMENT '返回页验证 verify验证 skip不验证',
  old_customer_days INT NOT NULL DEFAULT 30 COMMENT '老客户判断天数',
  pay_params TEXT NOT NULL COMMENT '支付参数',
  return_ip_whitelist TEXT NOT NULL COMMENT '返回验证IP白名单',
  pay_frequency INT NOT NULL DEFAULT 0 COMMENT '支付频率天',
  fail_count INT NOT NULL DEFAULT 0 COMMENT '失败次数限制',
  success_count INT NOT NULL DEFAULT 0 COMMENT '成功次数限制',
  fail_auto_close INT NOT NULL DEFAULT 0 COMMENT '失败自动关闭',
  mutual_hold_amount DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '互抛限制金额USD',
  amount_limit_mode VARCHAR(16) NOT NULL DEFAULT 'single' COMMENT '金额限制模式 single单笔 intercept拦截',
  calc_currency VARCHAR(16) NOT NULL DEFAULT 'USD' COMMENT '计算货币',
  allow_countries JSON NOT NULL COMMENT '仅支持国家',
  prefer_countries JSON NOT NULL COMMENT '优先国家',
  disable_countries JSON NOT NULL COMMENT '禁用国家',
  allow_card_types JSON NOT NULL COMMENT '仅支持卡类型',
  disable_card_types JSON NOT NULL COMMENT '禁用卡类型',
  disable_card_brands JSON NOT NULL COMMENT '禁用卡头',
  countries JSON NOT NULL COMMENT '限制国家',
  currencies JSON NOT NULL COMMENT '支付货币',
  mixers JSON NOT NULL COMMENT '一抛混流器',
  collect_rule VARCHAR(16) NOT NULL DEFAULT 'random' COMMENT '收款规则 random随机 round轮询',
  ship_range VARCHAR(32) NOT NULL DEFAULT '40-50' COMMENT '发货范围',
  sort INT NOT NULL DEFAULT 1 COMMENT '排序',
  auto_ship TINYINT(1) NOT NULL COMMENT '自动发货 1是 0否',
  return_keywords TEXT NOT NULL COMMENT '返回关键词',
  disable_brand_words TEXT NOT NULL COMMENT '禁用品牌词',
  created_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
  updated_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
  created_at DATETIME NULL COMMENT '创建时间',
  updated_at DATETIME NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY idx_channels_name (name),
  KEY idx_channels_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付通道配置表';

CREATE TABLE IF NOT EXISTS site_as (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'A站ID',
  merchant_id BIGINT UNSIGNED NOT NULL COMMENT '商户ID',
  domain VARCHAR(191) NOT NULL COMMENT '域名',
  framework VARCHAR(32) NOT NULL COMMENT '框架 woocommerce shopyy',
  status VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '状态 pending待审核 audited已审核 disabled禁用',
  created_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
  updated_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
  created_at DATETIME NULL COMMENT '创建时间',
  updated_at DATETIME NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY idx_site_as_domain (domain),
  KEY idx_site_as_merchant_id (merchant_id),
  KEY idx_site_as_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='A站管理表';

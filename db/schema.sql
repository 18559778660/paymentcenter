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

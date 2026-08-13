# Payment Center

这是 Go + Gin 支付中心最小骨架。

## 目标

- 接 A站请求
- 创建支付中心订单
- 预留路由选择
- 预留 `win_stripe` 接入点
- 后续再接 B站壳包和 Stripe 回调

## 当前功能

- 健康检查
- 创建订单
- 查询订单
- 标记成功
- 标记失败

## 目录结构

```text
paymentCenter/
├─ cmd/
│  └─ payment-center/
│     └─ main.go
├─ internal/
│  ├─ config/
│  │  └─ config.go
│  ├─ domain/
│  │  └─ order.go
│  ├─ store/
│  │  └─ memory.go
│  ├─ service/
│  │  └─ service.go
│  └─ transport/
│     └─ http/
│        └─ router.go
├─ docs/
│  └─ README.md
├─ go.mod
└─ go.sum
```

## 启动方式

```bash
go run ./cmd/payment-center
```

默认监听：

```text
:8080
```

可用环境变量覆盖：

```bash
set PAYMENT_CENTER_ADDR=:8081
```

建议在项目根目录放一个 `.env`，参考 `.env.example`。

### 常用环境变量

- `APP_ENV`
- `PAYMENT_CENTER_NAME`
- `PAYMENT_CENTER_ADDR`
- `STRIPE_API_KEY`
- `STRIPE_WEBHOOK_SECRET`

## 接口

### `GET /api/health`

返回服务状态。

### `POST /api/orders`

创建支付中心订单。

### `GET /api/orders`

查看全部订单。

### `GET /api/orders/:id`

查看单个订单。

### `POST /api/orders/:id/paid`

人工标记成功。

### `POST /api/orders/:id/failed`

人工标记失败。

## 统一响应

全局统一返回：

```json
{
  "code": 0,
  "data": {},
  "msg": "success"
}
```

约定：

- `code = 0` 成功
- `code = 1` 失败
- `data` 放业务数据
- `msg` 放提示信息

## 下一步建议

1. 加数据库层
2. 加商户表和密钥
3. 加通道路由表
4. 加 A站回调签名
5. 接 OpenCart `win_stripe` 或 `win`

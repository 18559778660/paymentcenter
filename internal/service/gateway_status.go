package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"paymentcenter/internal/model"
)

// 壳包 updateOrderStatus 回写的 status，与 function.php 注释一致：
// 0未付 1成功 2失败 3取消 4退款 5部分退款 6支付中
const (
	siteBOrderStatusUnpaid  = 0 // 未付
	siteBOrderStatusPaid    = 1 // 成功，return.php / notify.php
	siteBOrderStatusFailed  = 2 // 失败，return.php / notify.php
	siteBOrderStatusCancel  = 3 // 取消，cancel.php
	siteBOrderStatusRefund  = 4 // 退款，旧版主路不用
	siteBOrderStatusPartial = 5 // 部分退款，旧版主路不用
	siteBOrderStatusPaying  = 6 // 支付中，jump.php 去 Stripe 前
)

var (
	ErrGatewaySiteBTokenInvalid = errors.New("gateway site b token invalid")
	ErrGatewaySiteBOrderInvalid = errors.New("gateway site b order invalid")
)

// SiteBOrderStatusUpdateResult 壳包 orderStatusUpdate 回包。
type SiteBOrderStatusUpdateResult struct {
	RedirectNext string `json:"redirect_next"`
	Status       int64  `json:"status"`
}

type siteBOrderStatusPayload struct {
	Status            json.RawMessage `json:"status"`
	PayMsg            string          `json:"pay_msg"`
	SysOrderID        string          `json:"sys_order_id"`
	OrderNumber       string          `json:"order_number"`
	InvoiceNumber     string          `json:"invoiceNumber"`
	TransactionNumber string          `json:"transaction_number"`
}

// SiteBOrderStatusUpdate 接收 B 站壳包回写的支付状态，成功/失败时通知 A 站，并返回 redirect_next。
func (a *App) SiteBOrderStatusUpdate(token, orderJSON string) (SiteBOrderStatusUpdateResult, error) {
	var payload siteBOrderStatusPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(orderJSON)), &payload); err != nil {
		return SiteBOrderStatusUpdateResult{}, fmt.Errorf("%w: 订单参数无效", ErrGatewaySiteBOrderInvalid)
	}

	orderID := firstNonEmpty(strings.TrimSpace(payload.SysOrderID), strings.TrimSpace(payload.InvoiceNumber))
	if orderID == "" {
		return SiteBOrderStatusUpdateResult{}, fmt.Errorf("%w: 缺少 sys_order_id", ErrGatewaySiteBOrderInvalid)
	}
	if !a.verifySiteBAccessToken(orderID, token) {
		return SiteBOrderStatusUpdateResult{}, ErrGatewaySiteBTokenInvalid
	}

	order, err := a.store.GetOrder(orderID)
	if err != nil {
		if isNotFound(err) {
			return SiteBOrderStatusUpdateResult{}, fmt.Errorf("%w: 订单不存在", ErrGatewaySiteBOrderInvalid)
		}
		return SiteBOrderStatusUpdateResult{}, err
	}

	siteBStatus, err := parseFlexibleInt(payload.Status)
	if err != nil {
		return SiteBOrderStatusUpdateResult{}, fmt.Errorf("%w: 状态无效", ErrGatewaySiteBOrderInvalid)
	}

	changed, notifyA, err := a.applySiteBOrderStatus(order, siteBStatus, payload)
	if err != nil {
		return SiteBOrderStatusUpdateResult{}, err
	}
	if changed {
		order.UpdatedAt = time.Now().UTC()
		if err := a.store.SaveOrder(order); err != nil {
			return SiteBOrderStatusUpdateResult{}, err
		}
	}
	if notifyA {
		if err := a.notifyMerchantSite(order); err != nil {
			// 通知 A 失败不影响回给 B 的 redirect_next，否则顾客付完停在 B。
			a.appendOrderLogFromOrder(order, model.OrderLogTypeFailLog, "异步通知A站失败: "+err.Error())
		} else {
			a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "异步通知A站成功")
		}
	}

	return SiteBOrderStatusUpdateResult{
		RedirectNext: siteBRedirectNext(order, siteBStatus),
		Status:       siteBStatus,
	}, nil
}

// applySiteBOrderStatus 应用 B 站订单状态。
func (a *App) applySiteBOrderStatus(order *model.Order, siteBStatus int64, payload siteBOrderStatusPayload) (changed, notifyA bool, err error) {
	comment := strings.TrimSpace(payload.PayMsg)
	txn := strings.TrimSpace(payload.TransactionNumber)

	switch siteBStatus {
	case siteBOrderStatusPaying, siteBOrderStatusUnpaid:
		if order.Status == model.OrderStatusPaid {
			return false, false, nil
		}
		if order.Status == model.OrderStatusPending {
			if txn != "" && order.ProviderRef == "" {
				order.ProviderRef = txn
				return true, false, nil
			}
			return false, false, nil
		}
		order.Status = model.OrderStatusPending
		if txn != "" {
			order.ProviderRef = txn
		}
		a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "B站回写支付中，"+siteBWritebackLogDetail(comment))
		return true, false, nil

	case siteBOrderStatusPaid:
		alreadyPaid := order.Status == model.OrderStatusPaid
		changed := false
		if txn != "" && order.ProviderRef != txn {
			order.ProviderRef = txn
			changed = true
		}
		if alreadyPaid {
			a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "B站重复回写支付成功，忽略，"+siteBWritebackLogDetail(comment))
			return changed, false, nil
		}
		order.Status = model.OrderStatusPaid
		order.ErrorMessage = ""
		a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "B站回写支付成功，"+siteBWritebackLogDetail(comment))
		return true, true, nil

	case siteBOrderStatusFailed:
		if order.Status == model.OrderStatusPaid {
			a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "B站回写失败但订单已支付，忽略，"+siteBWritebackLogDetail(comment))
			return false, false, nil
		}
		if order.Status == model.OrderStatusFailed {
			return false, false, nil
		}
		order.Status = model.OrderStatusFailed
		order.ErrorMessage = firstNonEmpty(comment, "Payment failed")
		if txn != "" {
			order.ProviderRef = firstNonEmpty(order.ProviderRef, txn)
		}
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailRecord, "B站回写支付失败: "+order.ErrorMessage+"，"+siteBWritebackLogDetail(comment))
		return true, true, nil

	case siteBOrderStatusCancel:
		if order.Status == model.OrderStatusPaid {
			return false, false, nil
		}
		if order.Status == model.OrderStatusCancelled {
			return false, false, nil
		}
		order.Status = model.OrderStatusCancelled
		order.ErrorMessage = firstNonEmpty(comment, "Payment cancelled")
		if txn != "" {
			order.ProviderRef = firstNonEmpty(order.ProviderRef, txn)
		}
		a.appendOrderLogFromOrder(order, model.OrderLogTypeFailRecord, "B站回写支付取消，"+siteBWritebackLogDetail(comment))
		return true, true, nil

	case siteBOrderStatusRefund, siteBOrderStatusPartial:
		a.appendOrderLogFromOrder(order, model.OrderLogTypeRecord, "B站回写退款状态，当前版本暂不处理，"+siteBWritebackLogDetail(comment))
		return false, false, nil

	default:
		return false, false, fmt.Errorf("%w: 未知状态 %d", ErrGatewaySiteBOrderInvalid, siteBStatus)
	}
}

// siteBWritebackSource 根据壳包 pay_msg 区分回写入口。
// notify.php（Stripe webhook）comment 是 "notify:..."，return.php 是 "return completed..."。
func siteBWritebackSource(payMsg string) string {
	msg := strings.ToLower(strings.TrimSpace(payMsg))
	switch {
	case strings.HasPrefix(msg, "notify:"):
		return "webhook"
	case strings.HasPrefix(msg, "return"):
		return "返回页"
	case strings.Contains(msg, "payment redirect"):
		return "跳转"
	case msg == "cancel":
		return "取消页"
	default:
		return "未知"
	}
}

func siteBWritebackLogDetail(payMsg string) string {
	src := siteBWritebackSource(payMsg)
	msg := strings.TrimSpace(payMsg)
	if msg == "" {
		return "来源:" + src
	}
	return "来源:" + src + "，" + msg
}

// siteBRedirectNext 根据 B 站订单状态返回 redirect_next。
func siteBRedirectNext(order *model.Order, siteBStatus int64) string {
	switch siteBStatus {
	case siteBOrderStatusPaid:
		return firstNonEmpty(order.ResultURL, order.ReturnURL)
	case siteBOrderStatusFailed, siteBOrderStatusCancel:
		return firstNonEmpty(order.CancelURL, order.ReturnURL)
	default:
		return ""
	}
}

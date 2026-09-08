package service

import (
	"fmt"
	"strings"
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

// OrderLogListItem 管理端订单日志列表项。
type OrderLogListItem struct {
	ID            uint   `json:"id"`
	OrderID       string `json:"orderId"`
	MerchantID    uint   `json:"merchantId"`
	MerchantName  string `json:"merchantName"`
	MerchantOrder string `json:"orderNo"`
	Type          string `json:"type"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"createdAt"`
}

// OrderLogListQuery 订单日志列表筛选。
type OrderLogListQuery struct {
	MerchantID    uint
	MerchantOrder string
	CreatedFrom   string
	CreatedTo     string
}

// ListOrderLogs 查询订单日志列表。
func (a *App) ListOrderLogs(q OrderLogListQuery) ([]OrderLogListItem, error) {
	rows, err := a.store.ListOrderLogs(store.OrderLogListFilter{
		MerchantID:    q.MerchantID,
		MerchantOrder: strings.TrimSpace(q.MerchantOrder),
		CreatedFrom:   strings.TrimSpace(q.CreatedFrom),
		CreatedTo:     strings.TrimSpace(q.CreatedTo),
	})
	if err != nil {
		return nil, err
	}
	merchantIDs := make([]uint, 0, len(rows))
	seen := map[uint]struct{}{}
	for _, row := range rows {
		if row.MerchantID == 0 {
			continue
		}
		if _, ok := seen[row.MerchantID]; ok {
			continue
		}
		seen[row.MerchantID] = struct{}{}
		merchantIDs = append(merchantIDs, row.MerchantID)
	}
	nameMap := map[uint]string{}
	if len(merchantIDs) > 0 {
		merchants, err := a.store.GetMerchantsByIDs(merchantIDs)
		if err != nil {
			return nil, err
		}
		for _, m := range merchants {
			nameMap[m.ID] = m.Name
		}
	}
	out := make([]OrderLogListItem, 0, len(rows))
	for _, row := range rows {
		name := nameMap[row.MerchantID]
		if name == "" && row.MerchantID > 0 {
			name = fmt.Sprintf("#%d", row.MerchantID)
		}
		out = append(out, OrderLogListItem{
			ID:            row.ID,
			OrderID:       row.OrderID,
			MerchantID:    row.MerchantID,
			MerchantName:  name,
			MerchantOrder: row.MerchantOrder,
			Type:          row.Type,
			Remark:        row.Remark,
			CreatedAt:     row.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

// appendOrderLog 写入一条订单日志；写失败不影响主流程。
func (a *App) appendOrderLog(orderID string, merchantID uint, merchantOrder, logType, remark string) {
	remark = strings.TrimSpace(remark)
	if remark == "" {
		return
	}
	if logType == "" {
		logType = model.OrderLogTypeRecord
	}
	_ = a.store.CreateOrderLog(&model.OrderLog{
		OrderID:       strings.TrimSpace(orderID),
		MerchantID:    merchantID,
		MerchantOrder: strings.TrimSpace(merchantOrder),
		Type:          logType,
		Remark:        remark,
		CreatedAt:     time.Now().UTC(),
	})
}

// appendOrderLogFromOrder 按订单快照写日志。
func (a *App) appendOrderLogFromOrder(order *model.Order, logType, remark string) {
	if order == nil {
		return
	}
	a.appendOrderLog(order.ID, order.MerchantID, order.MerchantOrder, logType, remark)
}

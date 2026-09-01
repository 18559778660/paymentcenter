package service

import (
	"math/rand"
	"strings"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

// resolveGatewayRoute 解析网关路由。
func (a *App) resolveGatewayRoute(q GatewayPayQuery) (*model.Channel, *model.ChannelAccount, error) {
	channelName := strings.TrimSpace(q.Channel)
	groupCode := strings.TrimSpace(q.Group)
	if channelName == "" && groupCode == "" {
		return nil, nil, ErrGatewayRouteInvalid
	}
	if channelName != "" && groupCode != "" {
		return nil, nil, ErrGatewayRouteInvalid
	}
	if channelName != "" {
		channel, err := a.store.FindChannelByName(channelName)
		if err != nil {
			if isNotFound(err) {
				return nil, nil, ErrGatewayRouteInvalid
			}
			return nil, nil, err
		}
		account, err := a.pickChannelAccount(channel.ID, nil)
		if err != nil {
			return nil, nil, err
		}
		return channel, account, nil
	}
	group, err := a.store.FindChannelGroupByCode(groupCode)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, ErrGatewayRouteInvalid
		}
		return nil, nil, err
	}
	accountIDs, err := a.store.ListChannelGroupMemberAccountIDs(group.ID)
	if err != nil {
		return nil, nil, err
	}
	account, err := a.pickChannelAccount(0, accountIDs)
	if err != nil {
		return nil, nil, err
	}
	channel, err := a.store.GetChannelByID(account.ChannelID)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, ErrGatewayRouteInvalid
		}
		return nil, nil, err
	}
	return channel, account, nil
}

// pickChannelAccount 选择通道账号。
func (a *App) pickChannelAccount(channelID uint, accountIDs []uint) (*model.ChannelAccount, error) {
	var accounts []model.ChannelAccount
	var err error
	if channelID > 0 {
		cid := channelID
		accounts, err = a.store.ListChannelAccounts(store.ChannelAccountListFilter{
			ChannelID: &cid,
		})
	} else {
		all, listErr := a.store.ListChannelAccounts(store.ChannelAccountListFilter{})
		if listErr != nil {
			return nil, listErr
		}
		idSet := make(map[uint]struct{}, len(accountIDs))
		for _, id := range accountIDs {
			idSet[id] = struct{}{}
		}
		for _, item := range all {
			if _, ok := idSet[item.ID]; ok {
				accounts = append(accounts, item)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	enabled := make([]model.ChannelAccount, 0, len(accounts))
	for _, item := range accounts {
		if item.Status == model.ChannelAccountStatusEnabled {
			enabled = append(enabled, item)
		}
	}
	if len(enabled) == 0 {
		return nil, ErrGatewayAccountUnavailable
	}
	return &enabled[rand.Intn(len(enabled))], nil
}

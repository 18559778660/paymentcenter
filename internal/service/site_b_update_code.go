package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
)

var (
	ErrSiteBUpdateCodeIDsInvalid = errors.New("site b update code ids invalid")
)

var siteBHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// UpdateSiteBCodeRequest 批量更新代码。
type UpdateSiteBCodeRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// UpdateSiteBCodeResult 单个 B 站更新代码结果。
type UpdateSiteBCodeResult struct {
	ID      uint   `json:"id"`
	Domain  string `json:"domain"`
	OK      bool   `json:"ok"`
	Deploy  string `json:"deploy"`
	Webhook string `json:"webhook"`
	Message string `json:"message"`
}

// winMode1Response 壳包 top.php/win.php?mode=1 的返回。
type winMode1Response struct {
	Code json.RawMessage `json:"code"`
	Msg  string          `json:"msg"`
	Dir  string          `json:"dir"`
}

// siteBWebhookResponse 壳包 {运行目录}/stripe/webhook.php 的返回。
type siteBWebhookResponse struct {
	Code json.RawMessage `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

// UpdateSiteBCode 后台「更新代码」：对勾选的 B 站先部署壳包配置，再按绑定账号创建 Stripe webhook。
func (a *App) UpdateSiteBCode(req UpdateSiteBCodeRequest, operator string) ([]UpdateSiteBCodeResult, error) {
	ids := uniquePositiveIDs(req.IDs)
	if len(ids) == 0 {
		return nil, ErrSiteBUpdateCodeIDsInvalid
	}
	out := make([]UpdateSiteBCodeResult, 0, len(ids))
	for _, id := range ids {
		out = append(out, a.updateOneSiteBCode(id, operator))
	}
	return out, nil
}

// updateOneSiteBCode 处理单个 B 站：组 config → 调 mode=1 部署 → 调 webhook.php 创建回调。
func (a *App) updateOneSiteBCode(id uint, operator string) UpdateSiteBCodeResult {
	result := UpdateSiteBCodeResult{ID: id}
	siteB, err := a.store.GetSiteBByID(id)
	if err != nil {
		if isNotFound(err) {
			result.Message = "B站不存在"
			return result
		}
		result.Message = err.Error()
		return result
	}
	result.Domain = siteB.Domain

	platform, err := a.getPlatformByID(siteB.PlatformID)
	if err != nil {
		result.Message = "通道平台无效"
		return result
	}

	config, usedWords, err := a.buildSiteBShellConfig(siteB, platform)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	deployMsg, deployOK := a.deploySiteBShell(siteB.Domain, config)
	result.Deploy = deployMsg
	if deployOK && len(usedWords) > 0 {
		a.bumpStripeWordBankUsage(usedWords, operator)
	}

	webhookMsg, webhookOK := a.registerSiteBStripeWebhook(siteB, platform, config)
	result.Webhook = webhookMsg
	result.OK = deployOK && webhookOK
	result.Message = joinSiteBUpdateMessages(deployMsg, webhookMsg)
	return result
}

// buildSiteBShellConfig 组装 POST 给 mode=1 的 config.php 字段。
// dir 用后台运行目录；webhook_url / 各回调路径从单词库抽，抽完返回本次用过的词以便加使用次数。
func (a *App) buildSiteBShellConfig(siteB *model.SiteB, platform *model.Platform) (map[string]string, []model.StripeWordBank, error) {
	dir := strings.Trim(strings.TrimSpace(siteB.RunDirectory), "/")
	if dir == "" {
		dir = model.SiteBDefaultRunDirectory
	}
	used := map[string]struct{}{dir: {}}
	picked := make([]model.StripeWordBank, 0, 8)

	webhookURL, word, err := a.pickStripeWordBank(model.StripeWordBankConfigWebhook, used, "/webhook_page")
	if err != nil {
		return nil, nil, err
	}
	if word != nil {
		picked = append(picked, *word)
	}

	callbackKeys := []string{"form_url", "jump_url", "return_url", "cancel_url", "notify_url"}
	callbackPads := []string{"/form_page", "/jump_page", "/return_page", "/cancel_page", "/notify_page"}
	config := map[string]string{
		"dir":            dir,
		"channel":        strings.ToLower(strings.TrimSpace(platform.Code)),
		"invoice":        "invoice",
		"page":           "page",
		"order_no":       "order_no",
		"order_id":       "order_id",
		"webhook_url":    webhookURL,
		"invoice_prefix": "ORD",
		"invoice_length": "15",
		"invoice_tag":    "5",
		"type":           "0",
		"meta_data":      "",
		"refund_url":     "",
	}
	for i, key := range callbackKeys {
		value, word, err := a.pickStripeWordBank(model.StripeWordBankConfigCallback, used, callbackPads[i])
		if err != nil {
			return nil, nil, err
		}
		config[key] = value
		if word != nil {
			picked = append(picked, *word)
		}
	}
	return config, picked, nil
}

// pickStripeWordBank 从指定配置项（webhook链接 / 回调路径等）里抽一个未用过的词，优先用得少的；没有则用 fallback。
func (a *App) pickStripeWordBank(configItem string, used map[string]struct{}, fallback string) (string, *model.StripeWordBank, error) {
	list, err := a.store.ListStripeWordBanks(store.StripeWordBankListFilter{ConfigItem: configItem})
	if err != nil {
		return "", nil, err
	}
	sorted := append([]model.StripeWordBank(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].UsageCount != sorted[j].UsageCount {
			return sorted[i].UsageCount < sorted[j].UsageCount
		}
		return sorted[i].ID < sorted[j].ID
	})
	for i := range sorted {
		name := strings.TrimSpace(sorted[i].Name)
		if name == "" {
			continue
		}
		if _, exists := used[name]; exists {
			continue
		}
		if _, exists := used[strings.Trim(name, "/")]; exists {
			continue
		}
		used[name] = struct{}{}
		used[strings.Trim(name, "/")] = struct{}{}
		item := sorted[i]
		return name, &item, nil
	}
	fallback = strings.TrimSpace(fallback)
	used[fallback] = struct{}{}
	used[strings.Trim(fallback, "/")] = struct{}{}
	return fallback, nil, nil
}

// bumpStripeWordBankUsage 部署成功后给本次抽中的单词使用次数 +1。
func (a *App) bumpStripeWordBankUsage(items []model.StripeWordBank, operator string) {
	for i := range items {
		item := items[i]
		item.UsageCount++
		item.UpdatedBy = operator
		_ = a.store.SaveStripeWordBank(&item)
	}
}

// deploySiteBShell POST 整份 config 到 B 站 top.php?mode=1（没有再试 win.php）。
// 壳包会写入 stripe_deal/config.php 再改名为 config.dir；目标目录已存在则跳过改名。
func (a *App) deploySiteBShell(domain string, config map[string]string) (string, bool) {
	body, err := json.Marshal(config)
	if err != nil {
		return "组装配置失败: " + err.Error(), false
	}
	// 线上入口可能叫 top.php（win.php 改名），没有再试 win.php。
	entries := []struct {
		name string
		path string
	}{
		{name: "top.php", path: "/top.php?mode=1"},
		{name: "win.php", path: "/win.php?mode=1"},
	}
	lastErr := ""
	for _, entry := range entries {
		respBody, err := postSiteBJSON(siteBPublicURL(domain, entry.path), body)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		msg, ok, reachable := parseWinMode1Response(respBody, config["dir"])
		if !reachable {
			lastErr = msg
			continue
		}
		return entry.name + "：" + msg, ok
	}
	if lastErr == "" {
		lastErr = "top.php / win.php 均不可用"
	}
	return "部署请求失败: " + lastErr, false
}

// parseWinMode1Response 解析 mode=1 返回。code=1 为改名成功；「运行目录有误」视为已部署过、跳过改名。
func parseWinMode1Response(respBody []byte, fallbackDir string) (msg string, ok, reachable bool) {
	var raw winMode1Response
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "部署返回不是 JSON: " + truncateNotifyBody(respBody), false, false
	}
	code, _ := parseFlexibleInt(raw.Code)
	text := strings.TrimSpace(raw.Msg)
	if code == 1 {
		dir := strings.TrimSpace(raw.Dir)
		if dir == "" {
			dir = fallbackDir
		}
		return "已运行，目录 " + dir, true, true
	}
	if strings.Contains(text, "运行目录有误") {
		return "运行目录已存在，跳过改名", true, true
	}
	if text == "" {
		text = truncateNotifyBody(respBody)
	}
	return "部署失败: " + text, false, true
}

// registerSiteBStripeWebhook 对该 B 站每个已启用且有私钥的通道账号，请求壳包去 Stripe 登记 webhook。
func (a *App) registerSiteBStripeWebhook(siteB *model.SiteB, platform *model.Platform, config map[string]string) (string, bool) {
	if platform == nil || strings.ToLower(strings.TrimSpace(platform.Code)) != model.PlatformCodeStripe {
		return "非 Stripe，跳过 webhook", true
	}
	siteBID := siteB.ID
	accounts, err := a.store.ListChannelAccounts(store.ChannelAccountListFilter{SiteBID: &siteBID})
	if err != nil {
		return "查询通道账号失败: " + err.Error(), false
	}
	enabled := make([]model.ChannelAccount, 0, len(accounts))
	for _, account := range accounts {
		if account.Status != model.ChannelAccountStatusEnabled {
			continue
		}
		if strings.TrimSpace(account.PrivateKey) == "" {
			continue
		}
		enabled = append(enabled, account)
	}
	if len(enabled) == 0 {
		return "未绑定带私钥的通道账号，跳过 webhook", true
	}

	// 创建接口打正在跑的目录，不跟部署 body 里的 dir（避免改名目标尚未有 webhook.php）。
	dir := strings.Trim(strings.TrimSpace(siteB.RunDirectory), "/")
	if dir == "" {
		dir = model.SiteBDefaultRunDirectory
	}
	// webhookPath：Stripe 要 POST 的伪装路径，来自单词库 webhook链接，例如 /hook。
	webhookPath := strings.TrimSpace(config["webhook_url"])
	if !strings.HasPrefix(webhookPath, "/") {
		webhookPath = "/" + strings.TrimPrefix(webhookPath, "/")
	}
	callbackURL := siteBPublicURL(siteB.Domain, webhookPath)                  // 登记到 Stripe 的完整回调地址
	webhookAPI := siteBPublicURL(siteB.Domain, "/"+dir+"/stripe/webhook.php") // 壳包里真正调 Stripe API 的 PHP

	msgs := make([]string, 0, len(enabled))
	ok := true
	for i := range enabled {
		account := enabled[i]
		msg, createdOK := a.createStripeWebhookForAccount(&account, webhookAPI, callbackURL)
		msgs = append(msgs, fmt.Sprintf("账号%d %s", account.ID, msg))
		if !createdOK {
			ok = false
		}
	}
	return strings.Join(msgs, "；"), ok
}

// createStripeWebhookForAccount 用该账号私钥 sk_secret 调 B 站 webhook.php。
// 新建成功会返回 whsec_ 并写回通道账号；Stripe 上已有相同 callback_url 则不覆盖密钥。
func (a *App) createStripeWebhookForAccount(account *model.ChannelAccount, webhookAPI, callbackURL string) (string, bool) {
	payload, err := json.Marshal(map[string]string{
		"sk_secret":    strings.TrimSpace(account.PrivateKey),
		"callback_url": callbackURL,
	})
	if err != nil {
		return "组装 webhook 参数失败", false
	}
	respBody, err := postSiteBJSON(webhookAPI, payload)
	if err != nil {
		return "请求失败: " + err.Error(), false
	}
	var raw siteBWebhookResponse
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "返回不是 JSON: " + truncateNotifyBody(respBody), false
	}
	code, _ := parseFlexibleInt(raw.Code)
	dataText := rawJSONToTrimmedString(raw.Data)
	if code == 1 {
		secret := strings.TrimSpace(dataText)
		if strings.HasPrefix(secret, "whsec_") {
			account.WebSecret = secret
			if err := a.store.SaveChannelAccount(account); err != nil {
				return "已创建但保存密钥失败: " + err.Error(), false
			}
			return "已创建并保存 webhook 密钥", true
		}
		return "已创建但未返回 whsec", false
	}
	if strings.Contains(dataText, "已经创建过") {
		return "Stripe 上已有该回调", true
	}
	if dataText == "" {
		dataText = strings.TrimSpace(raw.Msg)
	}
	if dataText == "" {
		dataText = truncateNotifyBody(respBody)
	}
	return "创建失败: " + dataText, false
}

// postSiteBJSON POST JSON 到 B 站，忽略 TLS 证书校验。
func postSiteBJSON(target string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "payment-center")
	resp, err := siteBHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateNotifyBody(body))
	}
	return body, nil
}

// siteBPublicURL 把 B 站域名和路径拼成 https 地址。
func siteBPublicURL(domain, path string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Trim(domain, "/")
	if path == "" {
		return "https://" + domain
	}
	if strings.Contains(path, "?") {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return "https://" + domain + path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://" + domain + path
}

// uniquePositiveIDs 去掉 0 和重复的 B 站 ID，保持原顺序。
func uniquePositiveIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// joinSiteBUpdateMessages 把部署、webhook 等提示拼成一条给前端。
func joinSiteBUpdateMessages(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, "；")
}

// rawJSONToTrimmedString 把接口 data 字段（字符串或其它 JSON）收成纯文本。
func rawJSONToTrimmedString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.TrimSpace(asString)
	}
	return strings.TrimSpace(string(raw))
}

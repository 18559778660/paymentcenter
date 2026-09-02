package service

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
)

// shopyyGetSign 与 Shopyy 官方 getSign 一致：
// ksort → key=value&... → rtrim & → 拼接 &key=secret → MD5 大写。
func shopyyGetSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(params[key])
		builder.WriteByte('&')
	}
	stringA := strings.TrimSuffix(builder.String(), "&")
	toSign := stringA + "&key=" + secret

	sum := md5.Sum([]byte(toSign))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

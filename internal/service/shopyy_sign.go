package service

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

// shopyyGetSign 与 Shopyy getSign 一致。
func shopyyGetSign(data map[string]string, secret string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	str := ""
	for _, k := range keys {
		str += k + "=" + data[k] + "&"
	}
	stringA := strings.TrimSuffix(str, "&")
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(stringA+"&key="+secret))))
}

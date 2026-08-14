package store

import "errors"

// ErrNotFound 数据库层：按条件查询不到记录。
var ErrNotFound = errors.New("record not found")

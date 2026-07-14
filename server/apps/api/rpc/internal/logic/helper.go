package logic

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// isNotFound 判断错误是否为 sqlx 的“未找到记录”
func isNotFound(err error) bool {
	return errors.Is(err, sqlx.ErrNotFound)
}

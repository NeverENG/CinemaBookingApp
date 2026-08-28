package postgres

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

// DB 持有 gorm 连接，并向 repo 提供「事务优先」的查询入口。
// repo 不自己开事务：service 通过 TxManager 启动事务后，
// 事务通过 context 传递到这里，db() 自动优先返回事务。
type DB struct {
	gorm *gorm.DB
}

func NewDB(db *gorm.DB) *DB {
	return &DB{gorm: db}
}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

// db 返回 ctx 中的事务；没有事务时返回连接池。
func (d *DB) db(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return d.gorm
}

package postgres

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

// TxManager 实现 port.TxManager：service 启动事务，事务放入 ctx，
// repo 方法只接收 ctx，从 ctx 取出事务执行。
type TxManager struct {
	db *DB
}

var _ port.TxManager = (*TxManager)(nil)

func NewTxManager(db *DB) *TxManager {
	return &TxManager{db: db}
}

// Run 一个用例一个事务：fn 内的所有 repo 调用共享同一个 tx。
func (m *TxManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}

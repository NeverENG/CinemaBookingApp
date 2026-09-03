package postgresql

import (
	"context"

	"gorm.io/gorm"
)

type TxManager struct {
	db *gorm.DB
}

type TxKey struct{}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := tm.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err := tx.Error; err != nil {
			tx.Rollback()
		}
	}()

	Newctx := context.WithValue(ctx, TxKey{}, tx)

	if err := fn(Newctx); err != nil {
		return err
	}

	return tx.Commit().Error
}

package postgres

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type operationLogRow struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	AdminID    int64     `gorm:"column:admin_id"`
	Action     string    `gorm:"column:action"`
	TargetType string    `gorm:"column:target_type"`
	TargetID   string    `gorm:"column:target_id"`
	Detail     any       `gorm:"column:detail_json;serializer:json"`
	IP         string    `gorm:"column:ip"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (operationLogRow) TableName() string { return "operation_logs" }

// OperationLogRepo 实现 port.OperationLogRepo。
type OperationLogRepo struct {
	db *DB
}

var _ port.OperationLogRepo = (*OperationLogRepo)(nil)

func NewOperationLogRepo(db *DB) *OperationLogRepo {
	return &OperationLogRepo{db: db}
}

func (r *OperationLogRepo) Create(ctx context.Context, log *domain.OperationLog) error {
	detail := log.Detail
	// detail_json is NOT NULL; represent operations without extra details as
	// an empty JSON object so the audit write cannot abort the surrounding tx.
	if detail == nil {
		detail = map[string]any{}
	}
	row := operationLogRow{
		AdminID:    log.AdminID,
		Action:     log.Action,
		TargetType: log.TargetType,
		TargetID:   log.TargetID,
		Detail:     detail,
		IP:         log.IP,
		CreatedAt:  time.Now(),
	}
	return r.db.db(ctx).Create(&row).Error
}

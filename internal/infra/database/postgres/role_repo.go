package postgres

import (
	"context"
	"errors"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type roleRow struct {
	ID          int64    `gorm:"column:id;primaryKey"`
	Code        string   `gorm:"column:code"`
	Name        string   `gorm:"column:name"`
	Permissions []string `gorm:"column:permissions_json;serializer:json"`
	Status      string   `gorm:"column:status"`
}

func (roleRow) TableName() string { return "roles" }

// RoleRepo 实现 port.RoleRepo。
type RoleRepo struct {
	db *DB
}

var _ port.RoleRepo = (*RoleRepo)(nil)

func NewRoleRepo(db *DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	var row roleRow
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainRole(row), nil
}

func (r *RoleRepo) GetByCode(ctx context.Context, code string) (*domain.Role, error) {
	var row roleRow
	err := r.db.db(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainRole(row), nil
}

// Ensure 按 code 幂等补齐角色。
func (r *RoleRepo) Ensure(ctx context.Context, roles []domain.Role) error {
	for _, role := range roles {
		row := roleRow{Code: role.Code, Name: role.Name, Status: "ACTIVE"}
		if err := r.db.db(ctx).Where("code = ?", role.Code).FirstOrCreate(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func toDomainRole(row roleRow) *domain.Role {
	return &domain.Role{
		ID:          row.ID,
		Code:        row.Code,
		Name:        row.Name,
		Permissions: row.Permissions,
		Status:      row.Status,
	}
}

package postgres

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type seatRow struct {
	ID     int64  `gorm:"column:id;primaryKey"`
	HallID int64  `gorm:"column:hall_id"`
	RowNo  int    `gorm:"column:row_no"`
	ColNo  int    `gorm:"column:col_no"`
	SeatNo string `gorm:"column:seat_no"`
	Type   string `gorm:"column:type"`
	Status string `gorm:"column:status"`
}

func (seatRow) TableName() string { return "seats" }

// SeatRepo 实现 port.SeatRepo。
type SeatRepo struct {
	db *DB
}

var _ port.SeatRepo = (*SeatRepo)(nil)

func NewSeatRepo(db *DB) *SeatRepo {
	return &SeatRepo{db: db}
}

func (r *SeatRepo) ListSeatsByIDs(ctx context.Context, ids []int64) ([]domain.Seat, error) {
	if len(ids) == 0 {
		return []domain.Seat{}, nil
	}
	var rows []seatRow
	if err := r.db.db(ctx).Where("id IN ?", ids).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	seats := make([]domain.Seat, 0, len(rows))
	for _, row := range rows {
		seats = append(seats, domain.Seat{
			ID:     row.ID,
			HallID: row.HallID,
			RowNo:  row.RowNo,
			ColNo:  row.ColNo,
			SeatNo: row.SeatNo,
			Type:   row.Type,
			Status: row.Status,
		})
	}
	return seats, nil
}

func (r *SeatRepo) ListByHallID(ctx context.Context, hallID int64) ([]domain.Seat, error) {
	var rows []seatRow
	if err := r.db.db(ctx).Where("hall_id = ?", hallID).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	seats := make([]domain.Seat, 0, len(rows))
	for _, row := range rows {
		seats = append(seats, toDomainSeat(row))
	}
	return seats, nil
}

// SyncSeats 按布局 diff 同步：新增插入、变更更新、未出现的置 DISABLED（保留 ID 与历史）。
func (r *SeatRepo) SyncSeats(ctx context.Context, hallID int64, seats []domain.Seat) error {
	var existing []seatRow
	if err := r.db.db(ctx).Where("hall_id = ?", hallID).Find(&existing).Error; err != nil {
		return err
	}
	byNo := make(map[string]*seatRow, len(existing))
	for i := range existing {
		byNo[existing[i].SeatNo] = &existing[i]
	}

	desired := make(map[string]bool, len(seats))
	for _, s := range seats {
		desired[s.SeatNo] = true
		if row, ok := byNo[s.SeatNo]; ok {
			if row.Type != s.Type || row.Status != s.Status || row.RowNo != s.RowNo || row.ColNo != s.ColNo {
				row.Type, row.Status, row.RowNo, row.ColNo = s.Type, s.Status, s.RowNo, s.ColNo
				if err := r.db.db(ctx).Save(row).Error; err != nil {
					return err
				}
			}
			continue
		}
		if err := r.db.db(ctx).Create(&seatRow{
			HallID: hallID,
			RowNo:  s.RowNo,
			ColNo:  s.ColNo,
			SeatNo: s.SeatNo,
			Type:   s.Type,
			Status: s.Status,
		}).Error; err != nil {
			return err
		}
	}

	for no, row := range byNo {
		if !desired[no] && row.Status != domain.SeatDisabled {
			row.Status = domain.SeatDisabled
			if err := r.db.db(ctx).Save(row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func toDomainSeat(row seatRow) domain.Seat {
	return domain.Seat{
		ID:     row.ID,
		HallID: row.HallID,
		RowNo:  row.RowNo,
		ColNo:  row.ColNo,
		SeatNo: row.SeatNo,
		Type:   row.Type,
		Status: row.Status,
	}
}

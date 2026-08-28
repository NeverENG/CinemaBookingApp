package service

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// PointsSvc 用户积分查询。
type PointsSvc struct {
	points port.PointsRepo
}

func NewPointsSvc(points port.PointsRepo) *PointsSvc {
	return &PointsSvc{points: points}
}

type PointsView struct {
	Balance int                   `json:"balance"`
	Ledger  []domain.PointsLedger `json:"ledger"`
}

func (s *PointsSvc) GetPoints(ctx context.Context, userID int64) (*PointsView, error) {
	balance, err := s.points.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	ledger, err := s.points.GetRecentLedger(ctx, userID, 20)
	if err != nil {
		return nil, err
	}
	return &PointsView{Balance: balance, Ledger: ledger}, nil
}

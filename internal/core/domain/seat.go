package domain

// 座位物理状态。
const (
	SeatEnabled  = "ENABLED"
	SeatDisabled = "DISABLED"
)

// Seat 影厅座位。
type Seat struct {
	ID     int64
	HallID int64
	RowNo  int
	ColNo  int
	SeatNo string
	Type   string
	Status string
}

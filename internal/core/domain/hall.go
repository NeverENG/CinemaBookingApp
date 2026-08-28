package domain

type HallStatus string

const (
	HallActive      HallStatus = "ACTIVE"
	HallMaintenance HallStatus = "MAINTENANCE"
)

// Hall 影厅。SeatLayoutJSON 是前端渲染用的布局元数据，
// seats 表才是售票事实，保存时按布局 diff 同步。
type Hall struct {
	ID             int64
	CinemaID       int64
	Name           string
	SeatLayoutJSON string
	Status         HallStatus
}

func (h *Hall) Validate() error {
	if h.CinemaID <= 0 || h.Name == "" {
		return ErrHallInvalid
	}
	return nil
}

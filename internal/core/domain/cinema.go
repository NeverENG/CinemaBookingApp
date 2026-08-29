package domain

import "time"

type CinemaStatus string

const (
	CinemaActive   CinemaStatus = "ACTIVE"
	CinemaInactive CinemaStatus = "INACTIVE"
)

type Cinema struct {
	ID        int64
	Name      string
	City      string
	Address   string
	Longitude float64
	Latitude  float64
	Status    CinemaStatus
	CreatedAt time.Time
}

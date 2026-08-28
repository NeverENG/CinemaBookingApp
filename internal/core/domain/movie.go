package domain

import "time"

type MovieStatus string

const (
	MovieOnSale  MovieStatus = "ON_SALE"
	MovieOffSale MovieStatus = "OFF_SALE"
)

// Movie 影片（封面/预告片只存外部 URL）。
type Movie struct {
	ID              int64
	Title           string
	CoverURL        string
	TrailerURL      string
	Description     string
	DurationMinutes int
	Genre           string
	ReleaseDate     time.Time
	Rating          float64
	Status          MovieStatus
	CreatedAt       time.Time
}

func (m *Movie) Validate() error {
	if m.Title == "" || m.DurationMinutes <= 0 {
		return ErrMovieInvalid
	}
	if m.Rating < 0 || m.Rating > 10 {
		return ErrMovieInvalid
	}
	return nil
}

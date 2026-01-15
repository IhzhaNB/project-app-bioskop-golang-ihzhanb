package entity

import (
	"time"
)

type ReleaseStatus string

const (
	ReleaseStatusNowPlaying ReleaseStatus = "now_playing"
	ReleaseStatusComingSoon ReleaseStatus = "coming_soon"
)

type Movie struct {
	Base
	Title             string        `db:"title"`
	Description       *string       `db:"description"`
	PosterURL         *string       `db:"poster_url"`
	Rating            float64       `db:"rating"`
	ReleaseDate       time.Time     `db:"release_date"`
	DurationInMinutes int           `db:"duration_in_minutes"`
	ReleaseStatus     ReleaseStatus `db:"release_status"`
}

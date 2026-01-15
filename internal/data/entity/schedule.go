package entity

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	Base
	MovieID  uuid.UUID `db:"movie_id"`
	HallID   uuid.UUID `db:"hall_id"`
	ShowDate time.Time `db:"show_date"`
	ShowTime time.Time `db:"show_time"`
	Price    float64   `db:"price"`
}

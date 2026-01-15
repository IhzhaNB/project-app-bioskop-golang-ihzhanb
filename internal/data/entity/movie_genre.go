package entity

import (
	"github.com/google/uuid"
)

type MovieGenre struct {
	BaseSimple
	MovieID uuid.UUID `db:"movie_id"`
	GenreID uuid.UUID `db:"genre_id"`
}

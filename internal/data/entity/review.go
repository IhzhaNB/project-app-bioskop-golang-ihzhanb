package entity

import (
	"github.com/google/uuid"
)

type Review struct {
	BaseSimple
	UserID  uuid.UUID `db:"user_id"`
	MovieID uuid.UUID `db:"movie_id"`
	Rating  int       `db:"rating"` // 1-5
	Comment *string   `db:"comment"`
}

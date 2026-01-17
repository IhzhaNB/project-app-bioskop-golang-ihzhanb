package request

type CreateReviewRequest struct {
	MovieID string  `json:"movie_id" validate:"required,uuid4"`
	Rating  int     `json:"rating" validate:"required,min=1,max=5"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,max=500"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,max=500"`
}

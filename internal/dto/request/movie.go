package request

type MovieRequest struct {
	Title             string   `json:"title" validate:"required,min=1,max=200"`
	Description       *string  `json:"description,omitempty"`
	PosterURL         *string  `json:"poster_url,omitempty"`
	ReleaseDate       string   `json:"release_date" validate:"required,datetime=2006-01-02"`
	DurationInMinutes int      `json:"duration_in_minutes" validate:"required,min=1,max=999"`
	ReleaseStatus     string   `json:"release_status" validate:"required,oneof=now_playing coming_soon"`
	GenreIDs          []string `json:"genre_ids,omitempty" validate:"dive,uuid4"`
}

type MovieUpdateRequest struct {
	Title             *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description       *string `json:"description,omitempty"`
	PosterURL         *string `json:"poster_url,omitempty"`
	ReleaseDate       *string `json:"release_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	DurationInMinutes *int    `json:"duration_in_minutes,omitempty" validate:"omitempty,min=1,max=999"`
	ReleaseStatus     *string `json:"release_status,omitempty" validate:"omitempty,oneof=now_playing coming_soon"`
}

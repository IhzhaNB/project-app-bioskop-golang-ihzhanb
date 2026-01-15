package request

type GenreRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

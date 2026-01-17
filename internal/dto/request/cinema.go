package request

type CinemaRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Location string `json:"location" validate:"required,min=1,max=200"`
	City     string `json:"city" validate:"required,min=1,max=100"`
}

type CinemaUpdateRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Location *string `json:"location,omitempty" validate:"omitempty,min=1,max=200"`
	City     *string `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
}

type SeatAvailabilityRequest struct {
	Date string `json:"date" validate:"required,datetime=2006-01-02"`
	Time string `json:"time" validate:"required,datetime=15:04"`
}

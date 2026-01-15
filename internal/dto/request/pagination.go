package request

type PaginatedRequest struct {
	Page    int `json:"page" validate:"min=1"`
	PerPage int `json:"per_page" validate:"min=1,max=100"`
}

func (p PaginatedRequest) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

func (p PaginatedRequest) Limit() int {
	if p.PerPage < 1 {
		return 10
	}
	if p.PerPage > 100 {
		return 100
	}
	return p.PerPage
}

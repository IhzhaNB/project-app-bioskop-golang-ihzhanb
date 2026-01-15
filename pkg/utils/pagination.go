package utils

func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 || total <= 0 {
		return 0
	}
	return int((total + int64(perPage) - 1) / int64(perPage))
}

func CalculateOffset(page, perPage int) int {
	if page < 1 {
		return 0
	}
	return (page - 1) * perPage
}

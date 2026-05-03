package domain

import "math"

type Pagination struct {
	Page       int
	Limit      int
	TotalData  int64
	TotalPages int
}

func NewPagination(page, limit int, totalData int64) Pagination {
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	if totalPages == 0 && totalData > 0 {
		totalPages = 1
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		TotalData:  totalData,
		TotalPages: totalPages,
	}
}

type Paginated[T any] struct {
	Items      []T
	Pagination Pagination
}

type FindAllParams struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
}

package dto

import (
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/labstack/echo/v5"
)

type PaginationRequest struct {
	Page   int    `query:"page"   validate:"omitempty,min=1"          label:"Page"`
	Limit  int    `query:"limit"  validate:"omitempty,min=1"          label:"Limit"`
	Search string `query:"search" validate:"omitempty"                label:"Search"`
	Sort   string `query:"sort"   validate:"omitempty"                label:"Sort"`
	Order  string `query:"order"  validate:"omitempty,oneof=asc desc" label:"Order"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Status     int        `json:"status"`
	Message    string     `json:"message,omitempty"`
	Data       []T        `json:"data,omitempty"`
	Pagination Pagination `json:"pagination"`
}

func NewPaginatedResponse[T any](status int, data []T, p domain.Pagination, message ...string) PaginatedResponse[T] {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return PaginatedResponse[T]{
		Status:  status,
		Message: msg,
		Data:    data,
		Pagination: Pagination{
			Page:       p.Page,
			Limit:      p.Limit,
			TotalData:  p.TotalData,
			TotalPages: p.TotalPages,
		},
	}
}

func (r *PaginationRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.Order == "" {
		r.Order = "asc"
	}
}

func (r *PaginationRequest) ToDomain() domain.FindAllParams {
	return domain.FindAllParams{
		Page:   r.Page,
		Limit:  r.Limit,
		Search: r.Search,
		Sort:   r.Sort,
		Order:  r.Order,
	}
}

func (r *PaginationRequest) Bind(c *echo.Context) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	r.SetDefaults()
	return nil
}

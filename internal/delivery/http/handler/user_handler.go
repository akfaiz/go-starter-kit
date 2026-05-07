package handler

import (
	"errors"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *echo.Context) error {
	req := dto.UserListRequest{}
	if err := req.Bind(c); err != nil {
		return err
	}

	paginatedUsers, err := h.userService.FindAll(c.Request().Context(), req.ToDomain())
	if err != nil {
		return problem.Wrap(err, problem.ErrInternalServer)
	}

	userResponses := make([]*dto.UserResponse, len(paginatedUsers.Items))
	for i, u := range paginatedUsers.Items {
		userResponses[i] = dto.NewUserResponse(u)
	}

	res := dto.NewPaginatedResponse(
		200,
		userResponses,
		paginatedUsers.Pagination,
	)

	return c.JSON(res.Status, res)
}

func (h *UserHandler) GetUser(c *echo.Context) error {
	var req dto.UserGetRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	user, err := h.userService.FindByID(c.Request().Context(), req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return problem.ErrNotFound()
		}
		return problem.Wrap(err, problem.ErrInternalServer)
	}

	res := dto.NewResponse(200, dto.NewUserResponse(user))
	return c.JSON(res.Status, res)
}

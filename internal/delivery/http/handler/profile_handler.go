package handler

import (
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware/auth"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/errdefs"
	"github.com/labstack/echo/v5"
)

type ProfileHandler struct {
	userService domain.UserService
}

func NewProfileHandler(userService domain.UserService) *ProfileHandler {
	return &ProfileHandler{userService: userService}
}

func (h *ProfileHandler) GetProfile(c *echo.Context) error {
	claims := auth.GetUser(c)
	if claims == nil {
		return errdefs.ErrUnauthorized()
	}

	user, err := h.userService.FindByID(c.Request().Context(), claims.ID)
	if err != nil {
		return err
	}

	res := dto.NewResponse(200, dto.NewProfileResponse(user))
	return c.JSON(res.Status, res)
}

func (h *ProfileHandler) UpdateProfile(c *echo.Context) error {
	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := auth.GetUser(c)
	if claims == nil {
		return errdefs.ErrUnauthorized()
	}

	user, err := h.userService.FindByID(c.Request().Context(), claims.ID)
	if err != nil {
		return err
	}
	user.Name = req.Name
	user.Email = req.Email

	if err := h.userService.UpdateProfile(c.Request().Context(), claims.ID, user); err != nil {
		return err
	}

	updatedUser, err := h.userService.FindByID(c.Request().Context(), claims.ID)
	if err != nil {
		return err
	}

	res := dto.NewResponse(200, dto.NewProfileResponse(updatedUser))
	return c.JSON(res.Status, res)
}

func (h *ProfileHandler) ChangePassword(c *echo.Context) error {
	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := auth.GetUser(c)
	if claims == nil {
		return errdefs.ErrUnauthorized()
	}

	if err := h.userService.ChangePassword(
		c.Request().Context(),
		claims.ID,
		req.CurrentPassword,
		req.NewPassword,
	); err != nil {
		return err
	}

	res := dto.NewMessage(200, "Password changed successfully")
	return c.JSON(res.Status, res)
}

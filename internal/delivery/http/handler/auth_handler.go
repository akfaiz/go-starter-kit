package handler

import (
	"errors"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(
	authService domain.AuthService,
) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	pairToken, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			err = validator.NewError("email", i18n.T(c.Request().Context(), "auth.failed"))
		} else {
			err = problem.Wrap(err, problem.ErrInternalServer)
		}

		return err
	}

	res := dto.NewResponse(200, dto.NewTokenResponse(pairToken), "Login successful")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	token, err := h.authService.Register(c.Request().Context(), req.ToDomain())
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return validator.NewError("email", "Email already registered")
		}
		return problem.Wrap(err, problem.ErrInternalServer)
	}

	res := dto.NewResponse(201, dto.NewTokenResponse(token), "User registered successfully")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	pairToken, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return problem.ErrUnauthorized("invalid refresh token")
		}
		return problem.Wrap(err, problem.ErrInternalServer)
	}

	res := dto.NewResponse(200, dto.NewTokenResponse(pairToken))
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) SendForgotPasswordOTP(c *echo.Context) error {
	var req dto.SendForgotPasswordOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.authService.SendForgotPasswordOTP(c.Request().Context(), req.Email); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return validator.NewError("email", i18n.T(c.Request().Context(), "passwords.user"))
		}
		return problem.Wrap(err, problem.ErrInternalServer)
	}

	res := dto.NewMessage(200, i18n.T(c.Request().Context(), "passwords.sent"))
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) VerifyForgotPasswordOTP(c *echo.Context) error {
	var req dto.VerifyForgotPasswordOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.authService.VerifyForgotPasswordOTP(c.Request().Context(), req.Email, req.OTP); err != nil {
		return h.handleOTPError(c, err)
	}

	res := dto.NewMessage(200, "OTP is valid")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) ResetPasswordWithOTP(c *echo.Context) error {
	var req dto.ResetPasswordWithOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.authService.ResetPasswordWithOTP(c.Request().Context(), req.Email, req.OTP, req.Password); err != nil {
		return h.handleOTPError(c, err)
	}

	res := dto.NewMessage(200, "Password has been reset successfully")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) handleOTPError(c *echo.Context, err error) error {
	if errors.Is(err, domain.ErrUserNotFound) {
		return problem.ErrBadRequest(i18n.T(c.Request().Context(), "passwords.user"))
	}
	if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenExpired) {
		return problem.ErrBadRequest(i18n.T(c.Request().Context(), "passwords.token"))
	}
	return problem.Wrap(err, problem.ErrInternalServer)
}

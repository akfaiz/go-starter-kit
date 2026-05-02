package handler

import (
	"errors"
	"strconv"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/errdefs"
	"github.com/akfaiz/go-starter-kit/internal/validator"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authService domain.AuthService
	authGuard   domain.AuthGuard
}

func NewAuthHandler(authService domain.AuthService, authGuard domain.AuthGuard) *AuthHandler {
	return &AuthHandler{authService: authService, authGuard: authGuard}
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	check, err := h.authGuard.CheckLogin(c.Request().Context(), c.RealIP(), req.Email)
	if err != nil {
		return err
	}
	if check.Limited {
		return tooManyRequests(c, check.RetryAfter)
	}

	pairToken, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		var vErr *validator.ValidationError
		if errors.As(err, &vErr) && vErr.First().Field == "email" {
			limitResult, lErr := h.authGuard.OnLoginFailure(c.Request().Context(), c.RealIP(), req.Email)
			if lErr != nil {
				return lErr
			}
			if limitResult.Limited {
				return tooManyRequests(c, limitResult.RetryAfter)
			}
		}
		return err
	}
	if err := h.authGuard.OnLoginSuccess(c.Request().Context(), req.Email); err != nil {
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
	if err := c.Validate(&req); err != nil {
		return err
	}

	token, err := h.authService.Register(c.Request().Context(), req.ToDomain())
	if err != nil {
		return err
	}

	res := dto.NewResponse(201, dto.NewTokenResponse(token), "User registered successfully")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	check, err := h.authGuard.CheckRefresh(c.Request().Context(), c.RealIP())
	if err != nil {
		return err
	}
	if check.Limited {
		return tooManyRequests(c, check.RetryAfter)
	}

	pairToken, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return err
	}

	res := dto.NewResponse(200, dto.NewTokenResponse(pairToken))
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) SendForgotPasswordOTP(c *echo.Context) error {
	var req dto.SendForgotPasswordOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.SendForgotPasswordOTP(c.Request().Context(), req.Email); err != nil {
		return err
	}

	res := dto.NewMessage(200, i18n.T(c.Request().Context(), "passwords.sent"))
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) VerifyForgotPasswordOTP(c *echo.Context) error {
	var req dto.VerifyForgotPasswordOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.VerifyForgotPasswordOTP(c.Request().Context(), req.Email, req.OTP); err != nil {
		return err
	}

	res := dto.NewMessage(200, "OTP is valid")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) ResetPasswordWithOTP(c *echo.Context) error {
	var req dto.ResetPasswordWithOTPRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.ResetPasswordWithOTP(c.Request().Context(), req.Email, req.OTP, req.Password); err != nil {
		return err
	}

	res := dto.NewMessage(200, "Password has been reset successfully")
	return c.JSON(res.Status, res)
}

func tooManyRequests(c *echo.Context, retryAfter int) error {
	if retryAfter > 0 {
		c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))
	}
	return errdefs.ErrTooManyRequests("Too many authentication attempts. Please try again later.")
}

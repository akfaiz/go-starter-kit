package handler

import (
	"errors"
	"strconv"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/security"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authService domain.AuthService
	authGuard   security.AuthGuard
	validator   *validator.Validate
}

func NewAuthHandler(
	authService domain.AuthService,
	authGuard security.AuthGuard,
	validator *validator.Validate,
) *AuthHandler {
	return &AuthHandler{authService: authService, authGuard: authGuard, validator: validator}
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
		blockErr := h.handleLoginFailure(c, req.Email, err)
		if blockErr != nil {
			return blockErr
		}
		return err
	}
	if err := h.authGuard.OnLoginSuccess(c.Request().Context(), req.Email); err != nil {
		return err
	}

	res := dto.NewResponse(200, dto.NewTokenResponse(pairToken), "Login successful")
	return c.JSON(res.Status, res)
}

func (h *AuthHandler) handleLoginFailure(c *echo.Context, email string, loginErr error) error {
	var vErr *validator.ValidationError
	if !errors.As(loginErr, &vErr) || vErr.First().Field != "email" {
		return nil
	}

	limitResult, err := h.authGuard.OnLoginFailure(c.Request().Context(), c.RealIP(), email)
	if err != nil {
		return err
	}
	if limitResult.Limited {
		return tooManyRequests(c, limitResult.RetryAfter)
	}

	return nil
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
	if err := h.validator.ValidateContext(c.Request().Context(), &req); err != nil {
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
	return problem.ErrTooManyRequests("Too many authentication attempts. Please try again later.")
}

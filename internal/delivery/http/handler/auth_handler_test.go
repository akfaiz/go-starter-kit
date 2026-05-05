package handler_test

import (
	"context"
	"errors"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/test"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("AuthHandler", Label("unit", "handler"), func() {
	var (
		ctrl    *gomock.Controller
		service *mocks.MockAuthService
		h       *handler.AuthHandler
		e       *echo.Echo
		expect  *httpexpect.Expect
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		service = mocks.NewMockAuthService(ctrl)
		h = handler.NewAuthHandler(service)
		e = setupEcho()
		expect = newExpect(e)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Register", func() {
		BeforeEach(func() {
			e.POST("/auth/register", h.Register)
		})

		It("returns 201 with tokens when registration succeeds", func() {
			service.EXPECT().
				Register(gomock.Any(), gomock.AssignableToTypeOf(&domain.User{})).
				DoAndReturn(func(_ context.Context, user *domain.User) (*domain.PairToken, error) {
					Expect(user.Name).To(Equal("john"))
					Expect(user.Email).To(Equal("john@example.com"))
					Expect(user.Password).To(Equal("secret123"))
					return &domain.PairToken{AccessToken: "acc", RefreshToken: "ref"}, nil
				})

			expect.POST("/auth/register").
				WithJSON(map[string]any{
					"name":                  "john",
					"email":                 "john@example.com",
					"password":              "secret123",
					"password_confirmation": "secret123",
				}).
				Expect().
				Status(http.StatusCreated).
				JSON().
				Object().
				HasValue("status", 201).
				HasValue("message", "User registered successfully")
		})
	})

	Describe("Login", func() {
		BeforeEach(func() {
			e.POST("/auth/login", h.Login)
		})

		It("returns 200 with token response on success", func() {
			service.EXPECT().
				Login(gomock.Any(), "john@example.com", "secret123").
				Return(&domain.PairToken{AccessToken: "acc", RefreshToken: "ref"}, nil)

			expect.POST("/auth/login").
				WithJSON(map[string]any{
					"email":    "john@example.com",
					"password": "secret123",
				}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				HasValue("status", 200).
				HasValue("message", "Login successful")
		})

		It("returns 422 with error message when credentials are invalid", func() {
			service.EXPECT().
				Login(gomock.Any(), "john@example.com", "wrongpass").
				Return(nil, domain.ErrInvalidCredentials)

			expect.POST("/auth/login").
				WithJSON(map[string]any{
					"email":    "john@example.com",
					"password": "wrongpass",
				}).
				Expect().
				Status(http.StatusUnprocessableEntity).
				JSON(test.ProblemJSON).
				Object().
				Value("errors").
				Array().
				Value(0).
				Object().
				HasValue("message", "These credentials do not match our records")
		})
	})

	Describe("RefreshToken", func() {
		BeforeEach(func() {
			e.POST("/auth/refresh", h.RefreshToken)
		})

		It("returns 200 with new token pair", func() {
			service.EXPECT().
				RefreshToken(gomock.Any(), "r1").
				Return(&domain.PairToken{AccessToken: "new-acc", RefreshToken: "new-ref"}, nil)

			expect.POST("/auth/refresh").
				WithJSON(map[string]any{"refresh_token": "r1"}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				HasValue("status", 200)
		})
	})

	Describe("SendForgotPasswordOTP", func() {
		BeforeEach(func() {
			e.POST("/auth/forgot-password/send-otp", h.SendForgotPasswordOTP)
		})

		It("returns 200 when OTP send succeeds", func() {
			service.EXPECT().SendForgotPasswordOTP(gomock.Any(), "john@example.com").Return(nil)

			expect.POST("/auth/forgot-password/send-otp").
				WithJSON(map[string]any{"email": "john@example.com"}).
				Expect().
				Status(http.StatusOK)
		})
	})

	Describe("VerifyForgotPasswordOTP", func() {
		BeforeEach(func() {
			e.POST("/auth/forgot-password/verify-otp", h.VerifyForgotPasswordOTP)
		})

		It("returns 200 when OTP is valid", func() {
			service.EXPECT().VerifyForgotPasswordOTP(gomock.Any(), "john@example.com", "123456").Return(nil)

			expect.POST("/auth/forgot-password/verify-otp").
				WithJSON(map[string]any{
					"email": "john@example.com",
					"otp":   "123456",
				}).
				Expect().
				Status(http.StatusOK)
		})

		It("returns 400 when user is not found (handleOTPError)", func() {
			service.EXPECT().
				VerifyForgotPasswordOTP(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(domain.ErrUserNotFound)

			expect.POST("/auth/forgot-password/verify-otp").
				WithJSON(map[string]any{
					"email": "unknown@example.com",
					"otp":   "123456",
				}).
				Expect().
				Status(http.StatusBadRequest).
				JSON(test.ProblemJSON).
				Object().
				HasValue("title", "Bad Request")
		})

		It("returns 400 when token is invalid or expired (handleOTPError)", func() {
			service.EXPECT().
				VerifyForgotPasswordOTP(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(domain.ErrInvalidToken)

			expect.POST("/auth/forgot-password/verify-otp").
				WithJSON(map[string]any{
					"email": "john@example.com",
					"otp":   "111111",
				}).
				Expect().
				Status(http.StatusBadRequest).
				JSON(test.ProblemJSON)
		})

		It("returns 500 on unexpected error (handleOTPError)", func() {
			service.EXPECT().
				VerifyForgotPasswordOTP(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("db down"))

			expect.POST("/auth/forgot-password/verify-otp").
				WithJSON(map[string]any{
					"email": "john@example.com",
					"otp":   "123456",
				}).
				Expect().
				Status(http.StatusInternalServerError).
				JSON(test.ProblemJSON)
		})
	})

	Describe("ResetPasswordWithOTP", func() {
		BeforeEach(func() {
			e.POST("/auth/forgot-password/reset", h.ResetPasswordWithOTP)
		})

		It("returns 200 when password reset succeeds", func() {
			service.EXPECT().ResetPasswordWithOTP(gomock.Any(), "john@example.com", "123456", "newpassword").Return(nil)

			expect.POST("/auth/forgot-password/reset").
				WithJSON(map[string]any{
					"email":                 "john@example.com",
					"otp":                   "123456",
					"password":              "newpassword",
					"password_confirmation": "newpassword",
				}).
				Expect().
				Status(http.StatusOK)
		})
	})
})

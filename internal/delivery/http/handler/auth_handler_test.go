package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/security"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("AuthHandler", Label("unit", "handler"), func() {
	var (
		ctrl    *gomock.Controller
		service *mocks.MockAuthService
		guard   *mocks.MockAuthGuard
		h       *handler.AuthHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		service = mocks.NewMockAuthService(ctrl)
		guard = mocks.NewMockAuthGuard(ctrl)
		h = handler.NewAuthHandler(service, guard)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Register", func() {
		It("returns 201 with tokens when registration succeeds", func() {
			c, rec := newJSONContext(
				http.MethodPost,
				"/auth/register",
				`{"name":"john","email":"john@example.com","password":"secret123","password_confirmation":"secret123"}`,
			)

			service.EXPECT().
				Register(gomock.Any(), gomock.AssignableToTypeOf(&domain.User{})).
				DoAndReturn(func(_ context.Context, user *domain.User) (*domain.PairToken, error) {
					Expect(user.Name).To(Equal("john"))
					Expect(user.Email).To(Equal("john@example.com"))
					Expect(user.Password).To(Equal("secret123"))
					return &domain.PairToken{AccessToken: "acc", RefreshToken: "ref"}, nil
				})

			err := h.Register(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusCreated))

			var got map[string]any
			Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
			Expect(got["status"]).To(Equal(float64(201)))
			Expect(got["message"]).To(Equal("User registered successfully"))
		})
	})

	Describe("Login", func() {
		It("returns too many requests error and retry-after header when limited", func() {
			c, rec := newJSONContext(
				http.MethodPost,
				"/auth/login",
				`{"email":"john@example.com","password":"secret123"}`,
			)

			guard.EXPECT().
				CheckLogin(gomock.Any(), gomock.Any(), "john@example.com").
				Return(&security.RateLimitResult{Limited: true, RetryAfter: 30}, nil)

			err := h.Login(c)
			Expect(err).To(HaveOccurred())
			Expect(rec.Header().Get("Retry-After")).To(Equal("30"))

			var appErr *problem.AppError
			Expect(errors.As(err, &appErr)).To(BeTrue())
			Expect(appErr.Status).To(Equal(http.StatusTooManyRequests))
		})

		It("returns 200 with token response on success", func() {
			c, rec := newJSONContext(
				http.MethodPost,
				"/auth/login",
				`{"email":"john@example.com","password":"secret123"}`,
			)

			guard.EXPECT().
				CheckLogin(gomock.Any(), gomock.Any(), "john@example.com").
				Return(&security.RateLimitResult{}, nil)
			service.EXPECT().
				Login(gomock.Any(), "john@example.com", "secret123").
				Return(&domain.PairToken{AccessToken: "acc", RefreshToken: "ref"}, nil)
			guard.EXPECT().OnLoginSuccess(gomock.Any(), "john@example.com").Return(nil)

			err := h.Login(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var got map[string]any
			Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
			Expect(got["status"]).To(Equal(float64(200)))
			Expect(got["message"]).To(Equal("Login successful"))
		})
	})

	Describe("RefreshToken", func() {
		It("returns 200 with new token pair", func() {
			c, rec := newJSONContext(http.MethodPost, "/auth/refresh", `{"refresh_token":"r1"}`)

			guard.EXPECT().CheckRefresh(gomock.Any(), gomock.Any()).Return(&security.RateLimitResult{}, nil)
			service.EXPECT().
				RefreshToken(gomock.Any(), "r1").
				Return(&domain.PairToken{AccessToken: "new-acc", RefreshToken: "new-ref"}, nil)

			err := h.RefreshToken(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var got map[string]any
			Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
			Expect(got["status"]).To(Equal(float64(200)))
		})
	})

	Describe("SendForgotPasswordOTP", func() {
		It("returns 200 when OTP send succeeds", func() {
			c, rec := newJSONContext(http.MethodPost, "/auth/forgot-password/send-otp", `{"email":"john@example.com"}`)

			service.EXPECT().SendForgotPasswordOTP(gomock.Any(), "john@example.com").Return(nil)

			err := h.SendForgotPasswordOTP(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("VerifyForgotPasswordOTP", func() {
		It("returns 200 when OTP is valid", func() {
			c, rec := newJSONContext(
				http.MethodPost,
				"/auth/forgot-password/verify-otp",
				`{"email":"john@example.com","otp":"123456"}`,
			)

			service.EXPECT().VerifyForgotPasswordOTP(gomock.Any(), "john@example.com", "123456").Return(nil)

			err := h.VerifyForgotPasswordOTP(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("ResetPasswordWithOTP", func() {
		It("returns 200 when password reset succeeds", func() {
			c, rec := newJSONContext(
				http.MethodPost,
				"/auth/forgot-password/reset",
				`{"email":"john@example.com","otp":"123456","password":"newpassword","password_confirmation":"newpassword"}`,
			)

			service.EXPECT().ResetPasswordWithOTP(gomock.Any(), "john@example.com", "123456", "newpassword").Return(nil)

			err := h.ResetPasswordWithOTP(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})
})

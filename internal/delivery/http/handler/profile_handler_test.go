package handler_test

import (
	"net/http"
	"time"

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

var _ = Describe("ProfileHandler", Label("unit", "handler"), func() {
	var (
		ctrl        *gomock.Controller
		userService *mocks.MockUserService
		h           *handler.ProfileHandler
		e           *echo.Echo
		expect      *httpexpect.Expect
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userService = mocks.NewMockUserService(ctrl)
		h = handler.NewProfileHandler(userService)
		e = setupEcho()
		expect = newExpect(e)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	mockUserMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("user", &domain.JWTClaims{ID: 1})
			return next(c)
		}
	}

	Describe("GetProfile", func() {
		BeforeEach(func() {
			e.GET("/profile", h.GetProfile, mockUserMiddleware)
		})

		It("returns unauthorized when user claims are missing", func() {
			// Override route for this specific test case to remove middleware
			e = setupEcho()
			e.GET("/profile", h.GetProfile)
			expect = newExpect(e)

			expect.GET("/profile").
				Expect().
				Status(http.StatusUnauthorized).
				JSON(test.ProblemJSON).
				Object().
				HasValue("title", "Unauthorized access")
		})

		It("returns user profile when authenticated", func() {
			now := time.Now()
			userService.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.User{
				ID:        1,
				Name:      "John",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

			expect.GET("/profile").
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				Value("data").Object().
				HasValue("email", "john@example.com")
		})
	})

	Describe("UpdateProfile", func() {
		BeforeEach(func() {
			e.PUT("/profile", h.UpdateProfile, mockUserMiddleware)
		})

		It("updates and returns latest profile", func() {
			userService.EXPECT().
				FindByID(gomock.Any(), int64(1)).
				Return(&domain.User{ID: 1, Name: "John", Email: "john@example.com"}, nil)
			userService.EXPECT().
				UpdateProfile(gomock.Any(), int64(1), gomock.AssignableToTypeOf(&domain.User{})).
				DoAndReturn(
					func(_ any, _ int64, user *domain.User) error {
						Expect(user.Name).To(Equal("Jane"))
						Expect(user.Email).To(Equal("jane@example.com"))
						return nil
					},
				)
			userService.EXPECT().
				FindByID(gomock.Any(), int64(1)).
				Return(&domain.User{ID: 1, Name: "Jane", Email: "jane@example.com"}, nil)

			expect.PUT("/profile").
				WithJSON(map[string]any{"name": "Jane", "email": "jane@example.com"}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				Value("data").Object().
				HasValue("name", "Jane")
		})
	})

	Describe("ChangePassword", func() {
		BeforeEach(func() {
			e.PATCH("/profile/password", h.ChangePassword, mockUserMiddleware)
		})

		It("changes password and returns success message", func() {
			userService.EXPECT().ChangePassword(gomock.Any(), int64(1), "oldpass123", "newpass123").Return(nil)

			expect.PATCH("/profile/password").
				WithJSON(map[string]any{
					"current_password":          "oldpass123",
					"new_password":              "newpass123",
					"new_password_confirmation": "newpass123",
				}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				HasValue("message", "Password changed successfully")
		})
	})

	Describe("DeleteProfile", func() {
		BeforeEach(func() {
			e.DELETE("/profile", h.DeleteProfile, mockUserMiddleware)
		})

		It("deletes the profile and returns success message", func() {
			userService.EXPECT().Delete(gomock.Any(), int64(1), "delete-me").Return(nil)

			expect.DELETE("/profile").
				WithJSON(map[string]any{
					"current_password": "delete-me",
				}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				HasValue("message", "Profile deleted successfully")
		})

		It("returns validation error when current password is invalid", func() {
			userService.EXPECT().Delete(gomock.Any(), int64(1), "delete-me").Return(domain.ErrInvalidPassword)

			expect.DELETE("/profile").
				WithJSON(map[string]any{
					"current_password": "delete-me",
				}).
				Expect().
				Status(http.StatusUnprocessableEntity).
				JSON(test.ProblemJSON).
				Object().
				Value("errors").Array().
				Value(0).Object().
				HasValue("message", "Current password is incorrect")
		})
	})
})

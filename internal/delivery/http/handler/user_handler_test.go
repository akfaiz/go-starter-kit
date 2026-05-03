package handler_test

import (
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	"go.uber.org/mock/gomock"
)

var _ = Describe("UserHandler", Label("unit", "handler"), func() {
	var (
		ctrl        *gomock.Controller
		userService *mocks.MockUserService
		h           *handler.UserHandler
		e           *echo.Echo
		expect      *httpexpect.Expect
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userService = mocks.NewMockUserService(ctrl)
		h = handler.NewUserHandler(userService)
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

	Describe("ListUsers", func() {
		BeforeEach(func() {
			e.GET("/users", h.ListUsers, mockUserMiddleware)
		})

		It("returns users list with pagination", func() {
			userService.EXPECT().
				FindAll(gomock.Any(), domain.FindAllParams{Page: 1, Limit: 5, Order: "asc"}).
				Return(&domain.Paginated[*domain.User]{
					Items: []*domain.User{
						{ID: 1, Name: "User 1", Email: "user1@example.com"},
						{ID: 2, Name: "User 2", Email: "user2@example.com"},
					},
					Pagination: domain.Pagination{
						Page:       1,
						Limit:      5,
						TotalData:  2,
						TotalPages: 1,
					},
				}, nil)

			expect.GET("/users").
				WithQuery("page", 1).
				WithQuery("limit", 5).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object().
				HasValue("status", 200)
		})

		It("uses default values when pagination params are missing", func() {
			userService.EXPECT().
				FindAll(gomock.Any(), domain.FindAllParams{Page: 1, Limit: 10, Order: "asc"}).
				Return(&domain.Paginated[*domain.User]{
					Items: []*domain.User{},
					Pagination: domain.Pagination{
						Page:       1,
						Limit:      10,
						TotalData:  0,
						TotalPages: 0,
					},
				}, nil)

			expect.GET("/users").
				Expect().
				Status(http.StatusOK)
		})
	})
})

package handler_test

import (
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("UserHandler", Label("unit", "handler"), func() {
	var (
		ctrl        *gomock.Controller
		userService *mocks.MockUserService
		h           *handler.UserHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userService = mocks.NewMockUserService(ctrl)
		h = handler.NewUserHandler(userService)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("ListUsers", func() {
		It("returns users list with pagination", func() {
			c, rec := newJSONContext(http.MethodGet, "/users?page=1&limit=5", "")
			c.Set("user", &domain.JWTClaims{ID: 1})

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

			err := h.ListUsers(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})

		It("uses default values when pagination params are missing", func() {
			c, rec := newJSONContext(http.MethodGet, "/users", "")
			c.Set("user", &domain.JWTClaims{ID: 1})

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

			err := h.ListUsers(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})
})

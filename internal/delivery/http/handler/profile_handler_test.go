package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ProfileHandler", Label("unit", "handler"), func() {
	var (
		ctrl        *gomock.Controller
		userService *mocks.MockUserService
		h           *handler.ProfileHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userService = mocks.NewMockUserService(ctrl)
		h = handler.NewProfileHandler(userService, validator.New())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetProfile", func() {
		It("returns unauthorized when user claims are missing", func() {
			c, _ := newJSONContext(http.MethodGet, "/profile", "")

			err := h.GetProfile(c)
			Expect(err).To(HaveOccurred())

			var appErr *problem.AppError
			Expect(errors.As(err, &appErr)).To(BeTrue())
			Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
		})

		It("returns user profile when authenticated", func() {
			c, rec := newJSONContext(http.MethodGet, "/profile", "")
			c.Set("user", &domain.JWTClaims{ID: 1})

			now := time.Now()
			userService.EXPECT().FindByID(gomock.Any(), int64(1)).Return(&domain.User{
				ID:        1,
				Name:      "John",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

			err := h.GetProfile(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("UpdateProfile", func() {
		It("updates and returns latest profile", func() {
			c, rec := newJSONContext(http.MethodPut, "/profile", `{"name":"Jane","email":"jane@example.com"}`)
			c.Set("user", &domain.JWTClaims{ID: 1})

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

			err := h.UpdateProfile(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("ChangePassword", func() {
		It("changes password and returns success message", func() {
			c, rec := newJSONContext(
				http.MethodPatch,
				"/profile/password",
				`{"current_password":"oldpass123","new_password":"newpass123","new_password_confirmation":"newpass123"}`,
			)
			c.Set("user", &domain.JWTClaims{ID: 1})

			userService.EXPECT().ChangePassword(gomock.Any(), int64(1), "oldpass123", "newpass123").Return(nil)

			err := h.ChangePassword(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var got map[string]any
			Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
			Expect(got["message"]).To(Equal("Password changed successfully"))
		})
	})
})

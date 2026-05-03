package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	authmw "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware/auth"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/mocks"
	"github.com/akfaiz/go-starter-kit/pkg/errdefs"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Auth middleware", Label("unit", "middleware"), func() {
	var (
		ctrl *gomock.Controller
		jwt  *mocks.MockJWTManager
	)

	newContext := func(authHeader string) *echo.Context {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return c
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		jwt = mocks.NewMockJWTManager(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("returns unauthorized when Authorization header is missing", func() {
		c := newContext("")
		nextCalled := false

		err := authmw.New(jwt)(func(c *echo.Context) error {
			nextCalled = true
			return nil
		})(c)

		Expect(nextCalled).To(BeFalse())
		var appErr *errdefs.AppError
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns unauthorized when Authorization header is not Bearer", func() {
		c := newContext("Token abc")

		err := authmw.New(jwt)(func(c *echo.Context) error { return nil })(c)

		var appErr *errdefs.AppError
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns unauthorized when Bearer token is empty", func() {
		c := newContext("Bearer ")

		err := authmw.New(jwt)(func(c *echo.Context) error { return nil })(c)

		var appErr *errdefs.AppError
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("bubbles token verification error", func() {
		c := newContext("Bearer bad-token")
		jwt.EXPECT().VerifyAccessToken("bad-token").Return(nil, errors.New("invalid token"))

		err := authmw.New(jwt)(func(c *echo.Context) error { return nil })(c)
		Expect(err).To(MatchError("invalid token"))
	})

	It("sets claims in echo context and request context when token is valid", func() {
		c := newContext("Bearer good-token")
		claims := &domain.JWTClaims{ID: 1, Email: "john@example.com"}
		jwt.EXPECT().VerifyAccessToken("good-token").Return(claims, nil)

		nextCalled := false
		err := authmw.New(jwt)(func(c *echo.Context) error {
			nextCalled = true
			Expect(authmw.GetUser(c)).To(Equal(claims))
			Expect(authmw.GetUserFromContext(c.Request().Context())).To(Equal(claims))
			return nil
		})(c)

		Expect(err).NotTo(HaveOccurred())
		Expect(nextCalled).To(BeTrue())
	})
})

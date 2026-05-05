package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authmw "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware/auth"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Middleware Suite")
}

var _ = BeforeSuite(func() {
	lang.Init()
})

var _ = Describe("Auth middleware", Label("unit", "middleware"), func() {
	var (
		ctrl    *gomock.Controller
		jwt     *mocks.MockJWTManager
		session *mocks.MockSessionRepository
	)

	newContext := func(authHeader string) *echo.Context {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		ctx := req.Context()
		ctx, _ = ctxi18n.WithLocale(ctx, "en")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return c
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		jwt = mocks.NewMockJWTManager(ctrl)
		session = mocks.NewMockSessionRepository(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("returns unauthorized when Authorization header is missing", func() {
		c := newContext("")
		nextCalled := false

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error {
			nextCalled = true
			return nil
		})(c)

		Expect(nextCalled).To(BeFalse())
		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns localized unauthorized when Authorization header is missing (id)", func() {
		c := newContext("")
		// override locale to id
		ctx, _ := ctxi18n.WithLocale(c.Request().Context(), "id")
		c.SetRequest(c.Request().WithContext(ctx))

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
		Expect(appErr.Detail).To(Equal("header Authorization tidak ditemukan"))
	})

	It("returns unauthorized when Authorization header is not Bearer", func() {
		c := newContext("Token abc")

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns unauthorized when Bearer token is empty", func() {
		c := newContext("Bearer ")

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns unauthorized on token verification error", func() {
		c := newContext("Bearer bad-token")
		jwt.EXPECT().VerifyAccessToken("bad-token").Return(nil, errors.New("invalid token"))

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	It("returns token expired on expired token error", func() {
		c := newContext("Bearer expired-token")
		jwt.EXPECT().VerifyAccessToken("expired-token").Return(nil, domain.ErrTokenExpired)

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
		Expect(appErr.Title).To(Equal("Unauthorized"))
		Expect(appErr.Detail).To(Equal("Your session has expired. Please log in again."))
	})

	It("sets claims in echo context and request context when token is valid", func() {
		c := newContext("Bearer good-token")
		claims := &domain.JWTClaims{ID: 1, Email: "john@example.com"}
		jwt.EXPECT().VerifyAccessToken("good-token").Return(claims, nil)
		session.EXPECT().GetAccessToken(gomock.Any(), int64(1)).Return("good-token", nil)

		nextCalled := false
		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error {
			nextCalled = true
			Expect(authmw.GetUser(c)).To(Equal(claims))
			Expect(authmw.GetUserFromContext(c.Request().Context())).To(Equal(claims))
			return nil
		})(c)

		Expect(err).NotTo(HaveOccurred())
		Expect(nextCalled).To(BeTrue())
	})

	It("returns unauthorized when session token does not match", func() {
		c := newContext("Bearer good-token")
		claims := &domain.JWTClaims{ID: 1, Email: "john@example.com"}
		jwt.EXPECT().VerifyAccessToken("good-token").Return(claims, nil)
		session.EXPECT().GetAccessToken(gomock.Any(), int64(1)).Return("another-token", nil)

		err := authmw.NewWithSession(jwt, session)(func(c *echo.Context) error { return nil })(c)

		var appErr *problem.Error
		Expect(errors.As(err, &appErr)).To(BeTrue())
		Expect(appErr.Status).To(Equal(http.StatusUnauthorized))
	})

	Describe("New (without session)", func() {
		It("sets claims in echo context when token is valid", func() {
			c := newContext("Bearer good-token")
			claims := &domain.JWTClaims{ID: 1, Email: "john@example.com"}
			jwt.EXPECT().VerifyAccessToken("good-token").Return(claims, nil)

			nextCalled := false
			err := authmw.New(jwt)(func(c *echo.Context) error {
				nextCalled = true
				Expect(authmw.GetUser(c)).To(Equal(claims))
				return nil
			})(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(nextCalled).To(BeTrue())
		})
	})

	Describe("Helper functions", func() {
		It("GetUser returns nil when not set", func() {
			e := echo.New()
			c := e.NewContext(nil, nil)
			Expect(authmw.GetUser(c)).To(BeNil())
		})

		It("GetUserFromContext returns nil when not set", func() {
			Expect(authmw.GetUserFromContext(context.Background())).To(BeNil())
		})
	})
})

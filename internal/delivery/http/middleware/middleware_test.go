package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Middleware package", Label("unit", "middleware"), func() {
	It("I18n sets default locale when Accept-Language is missing", func() {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := middleware.I18n()(func(c *echo.Context) error {
			Expect(ctxi18n.Locale(c.Request().Context()).Code()).To(Equal(ctxi18n.DefaultLocale))
			return nil
		})(c)

		Expect(err).NotTo(HaveOccurred())
	})

	It("I18n uses Accept-Language when provided", func() {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := middleware.I18n()(func(c *echo.Context) error {
			Expect(string(ctxi18n.Locale(c.Request().Context()).Code())).To(HavePrefix("en"))
			return nil
		})(c)

		Expect(err).NotTo(HaveOccurred())
	})

	It("New wires auth middleware from JWT manager and session repository", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()

		jwt := mocks.NewMockJWTManager(ctrl)
		sessionRepo := mocks.NewMockSessionRepository(ctrl)
		out := middleware.New(middleware.Config{JWTManager: jwt, SessionRepo: sessionRepo})
		Expect(out.Auth).NotTo(BeNil())

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := out.Auth(func(c *echo.Context) error { return nil })(c)
		Expect(err).To(HaveOccurred())
	})
})

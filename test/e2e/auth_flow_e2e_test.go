package e2e_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Flow E2E", Label("e2e"), func() {
	BeforeEach(func() {
		_, err := e2eDB.ExecContext(e2eCtx, "TRUNCATE TABLE password_reset_tokens, users RESTART IDENTITY CASCADE")
		Expect(err).NotTo(HaveOccurred())
		Expect(e2eRDB.FlushDB(e2eCtx).Err()).NotTo(HaveOccurred())
	})

	It("returns healthy status", func() {
		e2eExpect.GET("/health").
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			Value("status").String().IsEqual("ok")
	})

	It("registers, logs in, and fetches profile successfully", func() {
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "John Doe",
				"email":                 "john@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated)

		loginObj := e2eExpect.POST("/api/v1/auth/login").
			WithJSON(map[string]any{
				"email":    "john@example.com",
				"password": "password123",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		accessToken := loginObj.Value("data").Object().Value("access_token").String().NotEmpty().Raw()

		profileObj := e2eExpect.GET("/api/v1/profile").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			Value("data").Object()
		profileObj.Value("email").String().IsEqual("john@example.com")
		profileObj.Value("name").String().IsEqual("John Doe")
	})

	It("returns validation error for invalid credentials", func() {
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "Jane Doe",
				"email":                 "jane@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated)

		resp := e2eExpect.POST("/api/v1/auth/login").
			WithJSON(map[string]any{
				"email":    "jane@example.com",
				"password": "wrong-password",
			}).
			Expect().
			Status(http.StatusUnprocessableEntity)
		resp.Header("Content-Type").Contains("application/problem+json")
		resp.Body().Contains(`"title":"Validation failed"`)
	})

	It("refreshes token successfully", func() {
		registerObj := e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "Refresh User",
				"email":                 "refresh@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().
			Object()

		refreshToken := registerObj.Value("data").Object().Value("refresh_token").String().NotEmpty().Raw()

		e2eExpect.POST("/api/v1/auth/refresh-token").
			WithJSON(map[string]any{
				"refresh_token": refreshToken,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			Value("data").Object().
			Value("access_token").String().NotEmpty()
	})

	It("rate limits repeated failed login attempts", func() {
		limited := false
		for range 6 {
			resp := e2eExpect.POST("/api/v1/auth/login").
				WithJSON(map[string]any{
					"email":    "locked@example.com",
					"password": "invalid-password",
				}).
				Expect()

			if resp.Raw().StatusCode == http.StatusTooManyRequests {
				resp.Header("Retry-After").NotEmpty()
				limited = true
				break
			}

			resp.Status(http.StatusUnprocessableEntity)
		}

		Expect(limited).To(BeTrue(), "expected 429 after repeated failed login attempts")
	})
})

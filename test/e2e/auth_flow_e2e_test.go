package e2e_test

import (
	"net/http"

	"github.com/akfaiz/go-starter-kit/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Flow E2E", Label("e2e"), func() {
	BeforeEach(func() {
		Expect(e2eDBContainer.TruncateAll(e2eCtx)).NotTo(HaveOccurred())
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

	It("deletes the profile successfully", func() {
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "Delete User",
				"email":                 "delete@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated)

		loginObj := e2eExpect.POST("/api/v1/auth/login").
			WithJSON(map[string]any{
				"email":    "delete@example.com",
				"password": "password123",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		accessToken := loginObj.Value("data").Object().Value("access_token").String().NotEmpty().Raw()

		e2eExpect.DELETE("/api/v1/profile").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(map[string]any{
				"current_password": "password123",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object().
			HasValue("message", "Profile deleted successfully")

		e2eExpect.POST("/api/v1/auth/login").
			WithJSON(map[string]any{
				"email":    "delete@example.com",
				"password": "password123",
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
		resp.JSON(test.ProblemJSON).Object().HasValue("title", "Validation failed")
		resp.JSON(test.ProblemJSON).
			Object().
			Value("errors").
			Array().
			Value(0).
			Object().
			HasValue("message", "These credentials do not match our records")
	})

	It("returns validation error for existing email during registration", func() {
		payload := map[string]any{
			"name":                  "Duplicate User",
			"email":                 "duplicate@example.com",
			"password":              "password123",
			"password_confirmation": "password123",
		}

		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(payload).
			Expect().
			Status(http.StatusCreated)

		resp := e2eExpect.POST("/api/v1/auth/register").
			WithJSON(payload).
			Expect().
			Status(http.StatusUnprocessableEntity)

		resp.JSON(test.ProblemJSON).
			Object().
			Value("errors").
			Array().
			Value(0).
			Object().
			HasValue("message", "Email already registered")
	})

	It("returns unauthorized when fetching profile without token", func() {
		e2eExpect.GET("/api/v1/profile").
			Expect().
			Status(http.StatusUnauthorized).
			JSON(test.ProblemJSON).
			Object().
			HasValue("title", "Unauthorized access")
	})

	It("returns unprocessable entity for malformed registration payload", func() {
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":  "Short",
				"email": "not-an-email",
			}).
			Expect().
			Status(http.StatusUnprocessableEntity).
			JSON(test.ProblemJSON).
			Object().
			HasValue("title", "Validation failed")
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
})

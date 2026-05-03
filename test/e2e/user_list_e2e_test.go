package e2e_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User List E2E", Label("e2e"), func() {
	BeforeEach(func() {
		Expect(e2eDBContainer.TruncateAll(e2eCtx)).NotTo(HaveOccurred())
		Expect(e2eRDB.FlushDB(e2eCtx).Err()).NotTo(HaveOccurred())
	})

	It("lists users with pagination and authentication", func() {
		// Register a user
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "Admin",
				"email":                 "admin@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated)

		// Register another user
		e2eExpect.POST("/api/v1/auth/register").
			WithJSON(map[string]any{
				"name":                  "User 1",
				"email":                 "user1@example.com",
				"password":              "password123",
				"password_confirmation": "password123",
			}).
			Expect().
			Status(http.StatusCreated)

		// Login as Admin
		loginObj := e2eExpect.POST("/api/v1/auth/login").
			WithJSON(map[string]any{
				"email":    "admin@example.com",
				"password": "password123",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		accessToken := loginObj.Value("data").Object().Value("access_token").String().Raw()

		// List users
		respObj := e2eExpect.GET("/api/v1/users").
			WithQuery("page", 1).
			WithQuery("limit", 10).
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		respObj.Value("data").Array().Length().IsEqual(2)
		respObj.Value("pagination").Object().Value("total_data").Number().IsEqual(2)
		respObj.Value("pagination").Object().Value("total_pages").Number().IsEqual(1)
		respObj.Value("pagination").Object().Value("page").Number().IsEqual(1)
		respObj.Value("pagination").Object().Value("limit").Number().IsEqual(10)
	})

	It("returns unauthorized when token is missing", func() {
		e2eExpect.GET("/api/v1/users").
			Expect().
			Status(http.StatusUnauthorized)
	})
})

package user_test

import (
	"context"
	"testing"

	"github.com/aarondl/opt/omit"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"github.com/akfaiz/go-starter-kit/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUserRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Repository Suite")
}

var (
	dbContainer *test.DBContainer
	r           domain.UserRepository
	ctx         context.Context
)

var _ = BeforeSuite(func() {
	ctx = context.Background()
	dbContainer = test.NewDBContainer(ctx, GinkgoT())
	Expect(dbContainer).NotTo(BeNil())
	r = user.NewRepository(dbContainer.DB)
})

var _ = AfterSuite(func() {
	if dbContainer != nil {
		err := dbContainer.Close(ctx)
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = Describe("User Repository", Label("unit", "repository", "integration"), func() {
	BeforeEach(func() {
		// Truncate all tables to ensure clean state for each test
		err := dbContainer.TruncateAll(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create", func() {
		var (
			u   *domain.User
			err error
		)

		BeforeEach(func() {
			u = &domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "hashed-password",
			}
		})

		JustBeforeEach(func() {
			err = r.Create(ctx, u)
		})

		When("successful", func() {
			It("should save the user", func() {
				Expect(err).To(BeNil())
			})

			It("should populate the user ID", func() {
				Expect(u.ID).NotTo(Equal(int64(0)))
			})

			It("should populate the created_at timestamp", func() {
				Expect(u.CreatedAt).NotTo(BeZero())
			})

			It("should populate the updated_at timestamp", func() {
				Expect(u.UpdatedAt).NotTo(BeZero())
			})
		})

		When("email already exists", func() {
			BeforeEach(func() {
				// Create first user
				firstUser := &domain.User{
					Name:     "Jane Doe",
					Email:    "john@example.com",
					Password: "hashed-password",
				}
				err := r.Create(ctx, firstUser)
				Expect(err).NotTo(HaveOccurred())

				// Try to create another user with same email
				u.Email = "john@example.com"
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(domain.ErrEmailAlreadyExists))
			})
		})
	})

	Describe("FindByEmail", func() {
		var (
			email string
			u     *domain.User
			err   error
		)

		BeforeEach(func() {
			email = "john@example.com"
		})

		JustBeforeEach(func() {
			u, err = r.FindByEmail(ctx, email)
		})

		When("user exists", func() {
			BeforeEach(func() {
				user := &domain.User{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "hashed-password",
				}
				err := r.Create(ctx, user)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the user", func() {
				Expect(err).To(BeNil())
				Expect(u).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(u.ID).NotTo(Equal(int64(0)))
			})

			It("should have correct name", func() {
				Expect(u.Name).To(Equal("John Doe"))
			})

			It("should have correct email", func() {
				Expect(u.Email).To(Equal("john@example.com"))
			})

			It("should have correct password", func() {
				Expect(u.Password).To(Equal("hashed-password"))
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				email = "missing@example.com"
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil user", func() {
				Expect(u).To(BeNil())
			})
		})
	})

	Describe("FindByID", func() {
		var (
			id  int64
			u   *domain.User
			err error
		)

		BeforeEach(func() {
			id = 1
		})

		JustBeforeEach(func() {
			u, err = r.FindByID(ctx, id)
		})

		When("user exists", func() {
			var createdUser *domain.User

			BeforeEach(func() {
				createdUser = &domain.User{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "hashed-password",
				}
				err := r.Create(ctx, createdUser)
				Expect(err).NotTo(HaveOccurred())
				id = createdUser.ID
			})

			It("should return the user", func() {
				Expect(err).To(BeNil())
				Expect(u).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(u.ID).To(Equal(id))
			})

			It("should have correct name", func() {
				Expect(u.Name).To(Equal("John Doe"))
			})

			It("should have correct email", func() {
				Expect(u.Email).To(Equal("john@example.com"))
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				id = 99999
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil user", func() {
				Expect(u).To(BeNil())
			})
		})
	})

	Describe("FindAll", func() {
		var (
			params    domain.FindAllParams
			paginated *domain.Paginated[*domain.User]
			err       error
		)

		BeforeEach(func() {
			params = domain.FindAllParams{
				Page:  1,
				Limit: 10,
			}

			// Create test data
			for i := 1; i <= 3; i++ {
				user := &domain.User{
					Name:     "User " + string(rune(48+i)),
					Email:    "user" + string(rune(48+i)) + "@example.com",
					Password: "hashed-password",
				}
				err := r.Create(ctx, user)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		JustBeforeEach(func() {
			paginated, err = r.FindAll(ctx, params)
		})

		When("successful", func() {
			It("should return users and total count", func() {
				Expect(err).To(BeNil())
				Expect(paginated.Items).To(HaveLen(3))
				Expect(paginated.Pagination.TotalData).To(Equal(int64(3)))
				Expect(paginated.Pagination.TotalPages).To(Equal(1))
			})
		})

		When("successful with search", func() {
			BeforeEach(func() {
				params.Search = "User 2"
			})

			It("should return filtered users", func() {
				Expect(err).To(BeNil())
				Expect(paginated.Items).To(HaveLen(1))
				Expect(paginated.Items[0].Name).To(ContainSubstring("User 2"))
			})
		})

		When("successful with sort", func() {
			BeforeEach(func() {
				params.Sort = "name"
				params.Order = "desc"
			})

			It("should return sorted users", func() {
				Expect(err).To(BeNil())
				Expect(paginated.Items).To(HaveLen(3))
				// Verify desc order: User 3, User 2, User 1
				Expect(paginated.Items[0].Name).To(Equal("User 3"))
				Expect(paginated.Items[2].Name).To(Equal("User 1"))
			})
		})

		When("pagination with limit", func() {
			BeforeEach(func() {
				params.Limit = 2
				params.Page = 1
			})

			It("should return limited users", func() {
				Expect(err).To(BeNil())
				Expect(paginated.Items).To(HaveLen(2))
				Expect(paginated.Pagination.TotalData).To(Equal(int64(3)))
				Expect(paginated.Pagination.TotalPages).To(Equal(2))
			})
		})
	})

	Describe("Update", func() {
		var (
			id         int64
			userUpdate *domain.UserUpdate
			err        error
		)

		BeforeEach(func() {
			// Create a user first
			user := &domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "hashed-password",
			}
			err := r.Create(ctx, user)
			Expect(err).NotTo(HaveOccurred())
			id = user.ID

			userUpdate = &domain.UserUpdate{}
		})

		JustBeforeEach(func() {
			err = r.Update(ctx, id, userUpdate)
		})

		When("successful update name", func() {
			BeforeEach(func() {
				userUpdate = &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should update the user", func() {
				updated, err := r.FindByID(ctx, id)
				Expect(err).NotTo(HaveOccurred())
				Expect(updated.Name).To(Equal("Jane Doe"))
			})
		})

		When("updating email to existing one", func() {
			BeforeEach(func() {
				// Create another user
				otherUser := &domain.User{
					Name:     "Other User",
					Email:    "other@example.com",
					Password: "hashed-password",
				}
				err := r.Create(ctx, otherUser)
				Expect(err).NotTo(HaveOccurred())

				// Try to update first user's email to the other user's email
				userUpdate = &domain.UserUpdate{
					Email: omit.From("other@example.com"),
				}
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				id = 99999
				userUpdate = &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})
	})

	Describe("Delete", func() {
		var (
			id  int64
			err error
		)

		BeforeEach(func() {
			// Create a user first
			user := &domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "hashed-password",
			}
			err := r.Create(ctx, user)
			Expect(err).NotTo(HaveOccurred())
			id = user.ID
		})

		JustBeforeEach(func() {
			err = r.Delete(ctx, id)
		})

		When("successful", func() {
			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should delete the user", func() {
				_, err := r.FindByID(ctx, id)
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				id = 99999
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})
	})
})

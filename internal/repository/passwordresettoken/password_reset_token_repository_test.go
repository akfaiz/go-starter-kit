package passwordresettoken_test

import (
	"context"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/repository/passwordresettoken"
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"github.com/akfaiz/go-starter-kit/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPasswordResetTokenRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Reset Token Repository Suite")
}

var (
	dbContainer *test.DBContainer
	r           domain.PasswordResetTokenRepository
	userRepo    domain.UserRepository
	ctx         context.Context
	testUserID  int64
)

var _ = BeforeSuite(func() {
	ctx = context.Background()
	dbContainer = test.NewDBContainer(ctx, GinkgoT())
	Expect(dbContainer).NotTo(BeNil())

	r = passwordresettoken.NewRepository(dbContainer.DB)
	userRepo = user.NewRepository(dbContainer.DB)
})

var _ = AfterSuite(func() {
	if dbContainer != nil {
		err := dbContainer.Close(ctx)
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = Describe("Password Reset Token Repository", Label("unit", "repository", "integration"), func() {
	BeforeEach(func() {
		// Truncate all tables to ensure clean state for each test
		err := dbContainer.TruncateAll(ctx)
		Expect(err).NotTo(HaveOccurred())

		// Recreate test user for foreign key references
		testUser := &domain.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "hashed-password",
		}
		err = userRepo.Create(ctx, testUser)
		Expect(err).NotTo(HaveOccurred())
		testUserID = testUser.ID
	})

	Describe("Create", func() {
		var (
			token *domain.PasswordResetToken
			err   error
		)

		When("successful", func() {
			BeforeEach(func() {
				token = &domain.PasswordResetToken{
					UserID:    testUserID,
					Token:     "test-token-123",
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
			})

			JustBeforeEach(func() {
				err = r.Create(ctx, token)
			})

			It("should save the token", func() {
				Expect(err).To(BeNil())
			})

			It("should populate the token ID", func() {
				Expect(token.ID).NotTo(Equal(int64(0)))
			})

			It("should populate the created_at timestamp", func() {
				Expect(token.CreatedAt).NotTo(BeZero())
			})
		})

		When("updating conflicting token", func() {
			var firstToken *domain.PasswordResetToken

			BeforeEach(func() {
				// Create first token for same user
				firstToken = &domain.PasswordResetToken{
					UserID:    testUserID,
					Token:     "old-token-123",
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				err := r.Create(ctx, firstToken)
				Expect(err).NotTo(HaveOccurred())

				// Create new token with same user (ON CONFLICT should update)
				token = &domain.PasswordResetToken{
					UserID:    testUserID,
					Token:     "new-token-456",
					ExpiresAt: time.Now().Add(2 * time.Hour),
				}
			})

			JustBeforeEach(func() {
				err = r.Create(ctx, token)
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("FindOne", func() {
		var (
			userID int64
			token  *domain.PasswordResetToken
			err    error
		)

		BeforeEach(func() {
			userID = testUserID
		})

		JustBeforeEach(func() {
			token, err = r.FindOne(ctx, userID)
		})

		When("token exists", func() {
			var createdToken *domain.PasswordResetToken

			BeforeEach(func() {
				createdToken = &domain.PasswordResetToken{
					UserID:    userID,
					Token:     "test-token-123",
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				err := r.Create(ctx, createdToken)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the token", func() {
				Expect(err).To(BeNil())
				Expect(token).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(token.ID).NotTo(Equal(int64(0)))
			})

			It("should have correct user ID", func() {
				Expect(token.UserID).To(Equal(userID))
			})

			It("should have correct token value", func() {
				Expect(token.Token).To(Equal("test-token-123"))
			})
		})

		When("token not found", func() {
			BeforeEach(func() {
				// Create a different user so we have tokens but not for our test user
				otherUser := &domain.User{
					Name:     "Other User",
					Email:    "other@example.com",
					Password: "hashed-password",
				}
				err := userRepo.Create(ctx, otherUser)
				Expect(err).NotTo(HaveOccurred())

				userID = 99999 // non-existent user
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil token", func() {
				Expect(token).To(BeNil())
			})
		})
	})

	Describe("Delete", func() {
		var (
			userID int64
			err    error
		)

		JustBeforeEach(func() {
			err = r.Delete(ctx, userID)
		})

		When("successful", func() {
			BeforeEach(func() {
				// Create a token first
				token := &domain.PasswordResetToken{
					UserID:    testUserID,
					Token:     "test-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				err := r.Create(ctx, token)
				Expect(err).NotTo(HaveOccurred())
				userID = testUserID
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should delete the token", func() {
				token, err := r.FindOne(ctx, userID)
				Expect(err).To(Equal(domain.ErrResourceNotFound))
				Expect(token).To(BeNil())
			})
		})

		When("no rows affected", func() {
			BeforeEach(func() {
				userID = 99999
			})

			It("should still succeed", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})

package passwordresettoken_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/repository/passwordresettoken"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
)

func TestPasswordResetTokenRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Reset Token Repository Suite")
}

var _ = Describe("Password Reset Token Repository", Label("unit", "repository"), func() {
	var (
		db   *bun.DB
		mock sqlmock.Sqlmock
		r    domain.PasswordResetTokenRepository
		ctx  context.Context
	)

	BeforeEach(func() {
		var cleanup func()
		db, mock, cleanup = mocks.NewMockDB(GinkgoT())

		DeferCleanup(func() {
			cleanup()
		})

		r = passwordresettoken.NewRepository(db)
		ctx = context.Background()
	})

	Describe("Create", func() {
		var (
			token *domain.PasswordResetToken
			err   error
		)

		When("successful", func() {
			BeforeEach(func() {
				now := time.Now()
				mock.ExpectQuery(`INSERT INTO "password_reset_tokens" AS "password_reset_token".*ON CONFLICT.*RETURNING`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

				token = &domain.PasswordResetToken{
					UserID:    1,
					Token:     "test-token-123",
					ExpiresAt: now.Add(1 * time.Hour),
				}
			})

			JustBeforeEach(func() {
				err = r.Create(ctx, token)
			})

			It("should save the token", func() {
				Expect(err).To(BeNil())
			})

			It("should populate the token ID", func() {
				Expect(token.ID).To(Equal(int64(1)))
			})

			It("should populate the created_at timestamp", func() {
				Expect(token.CreatedAt).NotTo(BeZero())
			})

			It("should verify all expectations were met", func() {
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("updating conflicting token", func() {
			BeforeEach(func() {
				now := time.Now()
				mock.ExpectQuery(`INSERT INTO "password_reset_tokens" AS "password_reset_token".*ON CONFLICT.*RETURNING`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

				token = &domain.PasswordResetToken{
					UserID:    1,
					Token:     "new-token-456",
					ExpiresAt: now.Add(1 * time.Hour),
				}
			})

			JustBeforeEach(func() {
				err = r.Create(ctx, token)
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`INSERT INTO "password_reset_tokens" AS "password_reset_token".*ON CONFLICT.*RETURNING`).
					WillReturnError(sql.ErrConnDone)

				token = &domain.PasswordResetToken{
					UserID:    1,
					Token:     "test-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
			})

			JustBeforeEach(func() {
				err = r.Create(ctx, token)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
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
			userID = 1
		})

		JustBeforeEach(func() {
			token, err = r.FindOne(ctx, userID)
		})

		When("token exists", func() {
			BeforeEach(func() {
				now := time.Now()
				expiresAt := now.Add(1 * time.Hour)

				mock.ExpectQuery(`SELECT "password_reset_token"."id", "password_reset_token"."user_id", "password_reset_token"."token", "password_reset_token"."expires_at", "password_reset_token"."created_at" FROM "password_reset_tokens"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
						AddRow(1, int64(1), "test-token-123", expiresAt, now))
			})

			It("should return the token", func() {
				Expect(err).To(BeNil())
				Expect(token).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(token.ID).To(Equal(int64(1)))
			})

			It("should have correct user ID", func() {
				Expect(token.UserID).To(Equal(int64(1)))
			})

			It("should have correct token value", func() {
				Expect(token.Token).To(Equal("test-token-123"))
			})
		})

		When("token not found", func() {
			BeforeEach(func() {
				userID = 99
				mock.ExpectQuery(`SELECT "password_reset_token"."id", "password_reset_token"."user_id", "password_reset_token"."token", "password_reset_token"."expires_at", "password_reset_token"."created_at" FROM "password_reset_tokens"`).
					WillReturnError(sql.ErrNoRows)
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil token", func() {
				Expect(token).To(BeNil())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT "password_reset_token"."id", "password_reset_token"."user_id", "password_reset_token"."token", "password_reset_token"."expires_at", "password_reset_token"."created_at" FROM "password_reset_tokens"`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
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
				userID = 1
				mock.ExpectExec(`DELETE FROM "password_reset_tokens" AS "password_reset_token" WHERE \(user_id = 1\)`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should verify all expectations were met", func() {
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("no rows affected", func() {
			BeforeEach(func() {
				userID = 99
				mock.ExpectExec(`DELETE FROM "password_reset_tokens" AS "password_reset_token" WHERE \(user_id = 99\)`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			})

			It("should still succeed", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})

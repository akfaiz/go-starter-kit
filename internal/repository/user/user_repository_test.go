package user_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aarondl/opt/omit"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
)

func TestUserRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Repository Suite")
}

var _ = Describe("User Repository", Label("unit", "repository"), func() {
	var (
		db   *bun.DB
		mock sqlmock.Sqlmock
		r    domain.UserRepository
		ctx  context.Context
	)

	BeforeEach(func() {
		var cleanup func()
		db, mock, cleanup = mocks.NewMockDB(GinkgoT())

		DeferCleanup(func() {
			cleanup()
		})

		r = user.NewRepository(db)
		ctx = context.Background()
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
			BeforeEach(func() {
				now := time.Now()
				mock.ExpectQuery(`INSERT INTO`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email_verified_at", "created_at", "updated_at"}).
						AddRow(1, nil, now, now))
			})

			It("should save the user", func() {
				Expect(err).To(BeNil())
			})

			It("should populate the user ID", func() {
				Expect(u.ID).To(Equal(int64(1)))
			})

			It("should populate the created_at timestamp", func() {
				Expect(u.CreatedAt).NotTo(BeZero())
			})

			It("should populate the updated_at timestamp", func() {
				Expect(u.UpdatedAt).NotTo(BeZero())
			})

			It("should verify all expectations were met", func() {
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("email already exists", func() {
			BeforeEach(func() {
				u.Email = "existing@example.com"
				// Simulating a database error; the actual pgdriver.Error would be caught in integration tests
				mock.ExpectQuery(`INSERT INTO`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return an error", func() {
				Expect(err).NotTo(BeNil())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`INSERT INTO`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
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
				now := time.Now()
				verifiedAt := now.Add(-24 * time.Hour)

				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "email_verified_at", "created_at", "updated_at"}).
						AddRow(1, "John Doe", "john@example.com", "hashed-password", verifiedAt, now, now))
			})

			It("should return the user", func() {
				Expect(err).To(BeNil())
				Expect(u).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(u.ID).To(Equal(int64(1)))
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
				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnError(sql.ErrNoRows)
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil user", func() {
				Expect(u).To(BeNil())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
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
			BeforeEach(func() {
				now := time.Now()
				verifiedAt := now.Add(-24 * time.Hour)

				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "email_verified_at", "created_at", "updated_at"}).
						AddRow(1, "John Doe", "john@example.com", "hashed-password", verifiedAt, now, now))
			})

			It("should return the user", func() {
				Expect(err).To(BeNil())
				Expect(u).NotTo(BeNil())
			})

			It("should have correct ID", func() {
				Expect(u.ID).To(Equal(int64(1)))
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
				id = 99
				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnError(sql.ErrNoRows)
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return nil user", func() {
				Expect(u).To(BeNil())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT .* FROM "users" AS "user" WHERE`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
			})

			It("should return nil user", func() {
				Expect(u).To(BeNil())
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
			id = 1
			userUpdate = &domain.UserUpdate{}
		})

		JustBeforeEach(func() {
			err = r.Update(ctx, id, userUpdate)
		})

		When("successful", func() {
			BeforeEach(func() {
				userUpdate = &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}
				mock.ExpectExec(`UPDATE "users" AS "user" SET`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should verify all expectations were met", func() {
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("updating email to existing one", func() {
			BeforeEach(func() {
				userUpdate = &domain.UserUpdate{
					Email: omit.From("existing@example.com"),
				}
				// Simulating a database error; the actual pgdriver.Error would be caught in integration tests
				mock.ExpectExec(`UPDATE "users"`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return an error", func() {
				Expect(err).NotTo(BeNil())
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				id = 99
				userUpdate = &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}
				mock.ExpectExec(`UPDATE "users" AS "user" SET`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				userUpdate = &domain.UserUpdate{
					Name: omit.From("Jane Doe"),
				}
				mock.ExpectExec(`UPDATE "users" AS "user" SET`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Delete", func() {
		var (
			id  int64
			err error
		)

		BeforeEach(func() {
			id = 1
		})

		JustBeforeEach(func() {
			err = r.Delete(ctx, id)
		})

		When("successful", func() {
			BeforeEach(func() {
				mock.ExpectExec(`DELETE FROM "users"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			})

			It("should succeed", func() {
				Expect(err).To(BeNil())
			})

			It("should verify all expectations were met", func() {
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				id = 99
				mock.ExpectExec(`DELETE FROM "users"`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			})

			It("should return resource not found error", func() {
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				mock.ExpectExec(`DELETE FROM "users"`).
					WillReturnError(sql.ErrConnDone)
			})

			It("should return error", func() {
				Expect(err).NotTo(BeNil())
			})
		})
	})
})

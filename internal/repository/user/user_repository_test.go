package user

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestFindByEmail_NoRowsReturnsResourceNotFound(t *testing.T) {
	db, mock, cleanup := mocks.NewMockDB(t)
	defer cleanup()

	r := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "user"."id", "user"."name", "user"."email", "user"."password", "user"."email_verified_at", "user"."created_at", "user"."updated_at" FROM "users" AS "user" WHERE (email = 'missing@example.com')`)).
		WillReturnError(sql.ErrNoRows)

	got, err := r.FindByEmail(context.Background(), "missing@example.com")
	assert.True(t, errors.Is(err, domain.ErrResourceNotFound))
	assert.Nil(t, got)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_NoRowsReturnsResourceNotFound(t *testing.T) {
	db, mock, cleanup := mocks.NewMockDB(t)
	defer cleanup()

	r := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" AS "user" WHERE ("user"."id" = 99)`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := r.Delete(context.Background(), 99)
	assert.True(t, errors.Is(err, domain.ErrResourceNotFound))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_Success(t *testing.T) {
	db, mock, cleanup := mocks.NewMockDB(t)
	defer cleanup()

	r := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" AS "user" WHERE ("user"."id" = 1)`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	assert.NoError(t, r.Delete(context.Background(), 1))

	assert.NoError(t, mock.ExpectationsWereMet())
}

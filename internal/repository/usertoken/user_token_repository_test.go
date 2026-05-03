package usertoken

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

func TestFindOne_NoRowsReturnsResourceNotFound(t *testing.T) {
	db, mock, cleanup := mocks.NewMockDB(t)
	defer cleanup()

	r := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "user_token"."id", "user_token"."user_id", "user_token"."token", "user_token"."token_type", "user_token"."expires_at", "user_token"."created_at" FROM "user_tokens" AS "user_token" WHERE (user_id = 10 AND token_type = 'refresh_token')`)).
		WillReturnError(sql.ErrNoRows)

	got, err := r.FindOne(context.Background(), 10, domain.TokenTypeRefreshToken)
	assert.True(t, errors.Is(err, domain.ErrResourceNotFound))
	assert.Nil(t, got)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_ExecutesQuery(t *testing.T) {
	db, mock, cleanup := mocks.NewMockDB(t)
	defer cleanup()

	r := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "user_tokens" AS "user_token" WHERE (user_id = 10 AND token_type = 'refresh_token')`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	assert.NoError(t, r.Delete(context.Background(), 10, domain.TokenTypeRefreshToken))

	assert.NoError(t, mock.ExpectationsWereMet())
}

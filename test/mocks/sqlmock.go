package mocks

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func NewMockDB(t ginkgo.GinkgoTInterface) (*bun.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqldb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	db := bun.NewDB(sqldb, pgdialect.New())
	cleanup := func() {
		_ = db.Close()
		_ = sqldb.Close()
	}

	return db, mock, cleanup
}

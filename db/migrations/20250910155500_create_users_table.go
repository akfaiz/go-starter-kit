package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(c schema.Context) error {
	return schema.Create(c, "users", func(table *schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("email").Unique()
		table.String("password")
		table.Timestamp("email_verified_at").Nullable()
		table.Timestamps()
	})
}

func downCreateUsersTable(c schema.Context) error {
	return schema.DropIfExists(c, "users")
}

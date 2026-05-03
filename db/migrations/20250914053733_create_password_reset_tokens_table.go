package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreatePasswordResetTokensTable, downCreatePasswordResetTokensTable)
}

func upCreatePasswordResetTokensTable(c schema.Context) error {
	return schema.Create(c, "password_reset_tokens", func(table *schema.Blueprint) {
		table.ID()
		table.BigInteger("user_id").Index()
		table.String("token")
		table.Timestamp("expires_at")
		table.Timestamp("created_at").UseCurrent()

		table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		table.Unique("user_id")
	})
}

func downCreatePasswordResetTokensTable(c schema.Context) error {
	return schema.DropIfExists(c, "password_reset_tokens")
}

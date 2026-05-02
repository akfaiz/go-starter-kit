package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateUserTokensTable, downCreateUserTokensTable)
}

func upCreateUserTokensTable(c schema.Context) error {
	return schema.Create(c, "user_tokens", func(table *schema.Blueprint) {
		table.ID()
		table.BigInteger("user_id").Index()
		table.Enum("token_type", []string{"forgot_password_otp", "refresh_token"})
		table.String("token")
		table.Timestamp("expires_at")
		table.Timestamp("created_at").UseCurrent()

		table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		table.Unique("user_id", "token_type")
	})
}

func downCreateUserTokensTable(c schema.Context) error {
	return schema.DropIfExists(c, "user_tokens")
}

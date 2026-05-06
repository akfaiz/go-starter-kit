package bcrypt_test

import (
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/hash/bcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	cfg := config.Config{
		Hash: config.Hash{
			BcryptCost: 10,
		},
	}
	hasher := bcrypt.NewHasher(cfg)

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "testpassword123"
		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
	})
}

func TestVerify(t *testing.T) {
	cfg := config.Config{
		Hash: config.Hash{
			BcryptCost: 10,
		},
	}
	hasher := bcrypt.NewHasher(cfg)

	t.Run("should verify correct password successfully", func(t *testing.T) {
		password := "testpassword123"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		valid, err := hasher.Verify(password, hash)

		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		valid, err := hasher.Verify(wrongPassword, hash)

		require.NoError(t, err)
		assert.False(t, valid)
	})
}

package argon2id_test

import (
	"strings"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/hash/argon2id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	cfg := config.Config{
		Hash: config.Hash{
			Argon2Memory:      64 * 1024,
			Argon2Iteration:   3,
			Argon2Parallelism: 1,
		},
	}
	hasher := argon2id.NewHasher(cfg)

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "testpassword123"

		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify hash format: $argon2id$v=19$m=65536,t=3,p=1$salt$hash
		parts := strings.Split(hash, "$")
		assert.Len(t, parts, 6)
		assert.Empty(t, parts[0]) // empty before first $
		assert.Equal(t, "argon2id", parts[1])
		assert.Equal(t, "v=19", parts[2])
		assert.Contains(t, hash, "$argon2id$")
	})

	t.Run("should handle special characters in password", func(t *testing.T) {
		password := "p@$$w0rd!@#$%^&*()_+-=[]{}|;':\",./<>?"

		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$argon2id$")
	})
}

func TestVerify(t *testing.T) {
	cfg := config.Config{
		Hash: config.Hash{
			Argon2Memory:      64 * 1024,
			Argon2Iteration:   3,
			Argon2Parallelism: 1,
		},
	}
	hasher := argon2id.NewHasher(cfg)

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

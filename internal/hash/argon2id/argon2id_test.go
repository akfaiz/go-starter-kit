package argon2id_test

import (
	"strings"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/hash/argon2id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	hasher := argon2id.NewHasher()

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "testpassword123"

		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify hash format: $argon2id$v=19$m=65536,t=3,p=1$salt$hash
		parts := strings.Split(hash, "$")
		assert.Len(t, parts, 6)
		assert.Equal(t, "", parts[0]) // empty before first $
		assert.Equal(t, "argon2id", parts[1])
		assert.Equal(t, "v=19", parts[2])
		assert.Equal(t, "m=65536,t=3,p=1", parts[3])
		assert.NotEmpty(t, parts[4]) // salt
		assert.NotEmpty(t, parts[5]) // hash
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := "samepassword"

		hash1, err1 := hasher.Hash(password)
		hash2, err2 := hasher.Hash(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("should handle empty password", func(t *testing.T) {
		password := ""

		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$argon2id$")
	})

	t.Run("should handle long password", func(t *testing.T) {
		password := strings.Repeat("a", 1000)

		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
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
	hasher := argon2id.NewHasher()

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

	t.Run("should handle empty password verification", func(t *testing.T) {
		password := ""
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		valid, err := hasher.Verify(password, hash)

		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("should reject invalid hash format - insufficient parts", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=19$m=65536,t=3,p=1$salt"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid password hash format")
	})

	t.Run("should reject invalid hash format - too many parts", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=19$m=65536,t=3,p=1$salt$hash$extra"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "invalid password hash format")
	})

	t.Run("should reject invalid version format", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=invalid$m=65536,t=3,p=1$c2FsdA$aGFzaA"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
	})

	t.Run("should reject unsupported argon2 version", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=18$m=65536,t=3,p=1$c2FsdA$aGFzaA"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "unsupported argon2 version")
	})

	t.Run("should reject invalid parameters format", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=19$m=invalid,t=3,p=1$c2FsdA$aGFzaA"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "failed to parse memory, iteration, or parallelism")
	})

	t.Run("should reject invalid salt encoding", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=19$m=65536,t=3,p=1$invalid@salt$aGFzaA"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "failed to decode salt")
	})

	t.Run("should reject invalid hash encoding", func(t *testing.T) {
		password := "testpassword"
		invalidHash := "$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$invalid@hash"

		valid, err := hasher.Verify(password, invalidHash)

		require.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "failed to decode hash")
	})

	t.Run("should verify password with special characters", func(t *testing.T) {
		password := "p@$$w0rd!@#$%^&*()_+-=[]{}|;':\",./<>?"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		valid, err := hasher.Verify(password, hash)

		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("should verify long password", func(t *testing.T) {
		password := strings.Repeat("a", 1000)
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		valid, err := hasher.Verify(password, hash)

		require.NoError(t, err)
		assert.True(t, valid)
	})
}

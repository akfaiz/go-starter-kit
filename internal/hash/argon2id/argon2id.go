package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/argon2"
)

type argon2idHasher struct {
	memory      uint32
	iteration   uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewHasher(cfg config.Config) domain.PasswordHasher {
	return &argon2idHasher{
		memory:      cfg.Hash.Argon2Memory,
		iteration:   cfg.Hash.Argon2Iteration,
		parallelism: cfg.Hash.Argon2Parallelism,
		saltLength:  16,
		keyLength:   32,
	}
}

func (h *argon2idHasher) Hash(password string) (string, error) {
	salt, err := h.generateSalt()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate salt")
	}

	hash := argon2.IDKey([]byte(password), salt, h.iteration, h.memory, h.parallelism, h.keyLength)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.iteration, h.parallelism, saltEncoded, hashEncoded,
	)

	return encodedHash, nil
}

func (h *argon2idHasher) Verify(password, passwordHashed string) (bool, error) {
	parts := strings.Split(passwordHashed, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid password hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("unsupported argon2 version")
	}

	var memory, iteration uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iteration, &parallelism)
	if err != nil {
		return false, errors.Wrap(err, "failed to parse memory, iteration, or parallelism")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errors.Wrap(err, "failed to decode salt")
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errors.Wrap(err, "failed to decode hash")
	}
	if uint64(len(hash)) > math.MaxUint32 {
		return false, errors.New("hash length overflows uint32")
	}
	//nolint:gosec // len(hash) comes from decoded digest; bounded by practical hash size.
	keyLength := uint32(len(hash))

	hashedPassword := argon2.IDKey([]byte(password), salt, iteration, memory, parallelism, keyLength)

	if subtle.ConstantTimeCompare(hashedPassword, hash) == 1 {
		return true, nil
	}

	return false, nil
}

func (h *argon2idHasher) generateSalt() ([]byte, error) {
	salt := make([]byte, h.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

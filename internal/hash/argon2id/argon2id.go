package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

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
	format      string
}

func NewHasher() domain.PasswordHasher {
	return &argon2idHasher{
		memory:      64 * 1024, // 64 MB
		iteration:   3,         // 3 iterations
		parallelism: 1,         // 1 parallel thread
		saltLength:  16,        // 16 bytes salt
		keyLength:   32,        // 32 bytes key length
		format:      "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
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

func (h *argon2idHasher) Verify(password, passwordHashed string) (valid bool, err error) {
	parts := strings.Split(passwordHashed, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid password hash format")
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return
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

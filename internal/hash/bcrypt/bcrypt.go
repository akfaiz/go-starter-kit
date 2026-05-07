package bcrypt

import (
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct {
	cost int
}

func NewHasher(cfg config.Config) domain.PasswordHasher {
	return &bcryptHasher{
		cost: cfg.Hash.BcryptCost,
	}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *bcryptHasher) Verify(password, passwordHashed string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(password))
	return err == nil, nil
}

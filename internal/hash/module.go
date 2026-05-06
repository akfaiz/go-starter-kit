package hash

import (
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/hash/argon2id"
	"github.com/akfaiz/go-starter-kit/internal/hash/bcrypt"
	"github.com/akfaiz/go-starter-kit/internal/hash/jwtmanager"
	"go.uber.org/fx"
)

func NewHasher(cfg config.Config) domain.PasswordHasher {
	if cfg.Hash.Driver == "bcrypt" {
		return bcrypt.NewHasher(cfg)
	}
	return argon2id.NewHasher(cfg)
}

var Module = fx.Module("hash",
	fx.Provide(
		NewHasher,
		jwtmanager.New,
	),
)

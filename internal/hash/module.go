package hash

import (
	"github.com/akfaiz/go-starter-kit/internal/hash/argon2id"
	"github.com/akfaiz/go-starter-kit/internal/hash/jwtmanager"
	"go.uber.org/fx"
)

var Module = fx.Module("hash",
	fx.Provide(
		argon2id.NewHasher,
		jwtmanager.New,
	),
)

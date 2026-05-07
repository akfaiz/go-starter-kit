package config

import (
	"math"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Hash struct {
	Driver string `validate:"required,oneof=argon2id bcrypt" label:"HASH_DRIVER"`

	// Argon2id settings
	Argon2Memory      uint32 `validate:"required,gt=0" label:"HASH_ARGON2_MEMORY"`
	Argon2Iteration   uint32 `validate:"required,gt=0" label:"HASH_ARGON2_ITERATION"`
	Argon2Parallelism uint8  `validate:"required,gt=0" label:"HASH_ARGON2_PARALLELISM"`

	// Bcrypt settings
	BcryptCost int `validate:"required,min=4,max=31" label:"HASH_BCRYPT_COST"`
}

func loadHashConfig() Hash {
	return Hash{
		Driver:            env.GetString("HASH_DRIVER", "argon2id"),
		Argon2Memory:      sanitizeUint32(env.GetInt("HASH_ARGON2_MEMORY", 64*1024)),
		Argon2Iteration:   sanitizeUint32(env.GetInt("HASH_ARGON2_ITERATION", 3)),
		Argon2Parallelism: sanitizeUint8(env.GetInt("HASH_ARGON2_PARALLELISM", 1)),
		BcryptCost:        env.GetInt("HASH_BCRYPT_COST", 10),
	}
}

func sanitizeUint32(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}

func sanitizeUint8(value int) uint8 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(value)
}

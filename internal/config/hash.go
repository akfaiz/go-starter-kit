package config

import "github.com/akfaiz/go-starter-kit/pkg/env"

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
		Argon2Memory:      uint32(env.GetInt("HASH_ARGON2_MEMORY", 64*1024)),
		Argon2Iteration:   uint32(env.GetInt("HASH_ARGON2_ITERATION", 3)),
		Argon2Parallelism: uint8(env.GetInt("HASH_ARGON2_PARALLELISM", 1)),
		BcryptCost:        env.GetInt("HASH_BCRYPT_COST", 10),
	}
}

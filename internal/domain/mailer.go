//go:generate mockgen -source=mailer.go -destination=../mocks/mailer_mock.go -package=mocks
package domain

import (
	"context"

	"github.com/akfaiz/go-mailgen"
)

type Mailer interface {
	Send(ctx context.Context, msg *mailgen.Builder) error
}

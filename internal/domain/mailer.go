//go:generate mockgen -source=mailer.go -destination=../../test/mocks/mailer_mock.go -package=mocks
package domain

import (
	"context"
)

type Mail struct {
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Text    string
	HTML    string
}

type Mailer interface {
	Send(ctx context.Context, mail *Mail) error
}

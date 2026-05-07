package infra_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPMailer(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Name:            "TestApp",
			FrontendBaseURL: "http://localhost",
		},
		Mail: config.Mail{
			SMTP: config.MailSMTP{
				Host:     "localhost",
				Port:     587,
				Username: "user",
				Password: "password",
				TLSMode:  "none",
			},
		},
	}

	mailer, err := infra.NewSMTPMailer(cfg)
	require.NoError(t, err)
	require.NotNil(t, mailer)

	t.Run("TLS mode none", func(t *testing.T) {
		cfg.Mail.SMTP.TLSMode = "none"
		m, err := infra.NewSMTPMailer(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("TLS mode tls", func(t *testing.T) {
		cfg.Mail.SMTP.TLSMode = "tls"
		m, err := infra.NewSMTPMailer(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("TLS mode mandatory", func(t *testing.T) {
		cfg.Mail.SMTP.TLSMode = "mandatory"
		m, err := infra.NewSMTPMailer(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestNewMailer_DriverSelection(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Name:            "TestApp",
			FrontendBaseURL: "http://localhost",
		},
		Mail: config.Mail{
			Driver: "smtp",
			SMTP: config.MailSMTP{
				Host:    "localhost",
				Port:    587,
				TLSMode: "none",
			},
		},
	}

	t.Run("smtp", func(t *testing.T) {
		m, err := infra.NewMailer(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("log", func(t *testing.T) {
		cfg.Mail.Driver = "log"
		m, err := infra.NewMailer(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("invalid", func(t *testing.T) {
		cfg.Mail.Driver = "invalid"
		m, err := infra.NewMailer(cfg)
		assert.Nil(t, m)
		assert.Error(t, err)
	})
}

func TestSMTPMailer_Send_Validation(t *testing.T) {
	cfg := config.Config{
		App: config.App{Name: "TestApp"},
		Mail: config.Mail{
			SMTP: config.MailSMTP{Host: "localhost", Port: 587},
			From: config.MailFrom{Name: "Sender", Address: "sender@example.com"},
		},
	}
	mailer, _ := infra.NewSMTPMailer(cfg)
	ctx := context.Background()

	t.Run("nil builder", func(t *testing.T) {
		err := mailer.Send(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mail message cannot be nil")
	})

	t.Run("no recipient", func(t *testing.T) {
		msg := &domain.Mail{Subject: "Test"}
		err := mailer.Send(ctx, msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email recipient cannot be empty")
	})

	t.Run("no subject", func(t *testing.T) {
		msg := &domain.Mail{To: []string{"test@example.com"}}
		err := mailer.Send(ctx, msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email subject cannot be empty")
	})
}

func TestLogMailer_Send(t *testing.T) {
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })
	slog.SetDefault(slog.New(slog.DiscardHandler))

	cfg := config.Config{Mail: config.Mail{Driver: "log"}}
	mailer, err := infra.NewMailer(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("nil message", func(t *testing.T) {
		err := mailer.Send(ctx, nil)
		assert.EqualError(t, err, "mail message cannot be nil")
	})

	t.Run("empty to", func(t *testing.T) {
		err := mailer.Send(ctx, &domain.Mail{Subject: "subject"})
		assert.EqualError(t, err, "email recipient cannot be empty")
	})

	t.Run("empty subject", func(t *testing.T) {
		err := mailer.Send(ctx, &domain.Mail{To: []string{"to@example.com"}})
		assert.EqualError(t, err, "email subject cannot be empty")
	})

	t.Run("success", func(t *testing.T) {
		err := mailer.Send(ctx, &domain.Mail{
			To:      []string{"to@example.com"},
			Cc:      []string{"cc@example.com"},
			Bcc:     []string{"bcc@example.com"},
			Subject: "subject",
			Text:    "plain",
			HTML:    "<b>html</b>",
		})
		assert.NoError(t, err)
	})
}

package infra_test

import (
	"context"
	"testing"

	"github.com/akfaiz/go-mailgen"
	"github.com/akfaiz/go-starter-kit/internal/config"
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
		assert.Contains(t, err.Error(), "message builder cannot be nil")
	})

	t.Run("no recipient", func(t *testing.T) {
		builder := mailgen.New().Subject("Test")
		err := mailer.Send(ctx, builder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email recipient cannot be empty")
	})

	t.Run("no subject", func(t *testing.T) {
		builder := mailgen.New().To("test@example.com")
		err := mailer.Send(ctx, builder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email subject cannot be empty")
	})
}

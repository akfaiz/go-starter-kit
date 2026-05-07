package infra

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/akfaiz/go-mailgen"
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	cerrors "github.com/cockroachdb/errors"
	"github.com/wneessen/go-mail"
)

type smtpMailer struct {
	client  *mail.Client
	mailCfg config.Mail
}

type logMailer struct{}

func NewMailer(cfg config.Config) (domain.Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Mail.Driver)) {
	case "log":
		return &logMailer{}, nil
	case "smtp":
		return NewSMTPMailer(cfg)
	default:
		return nil, cerrors.WithStack(fmt.Errorf("unsupported mail driver: %s", cfg.Mail.Driver))
	}
}

func NewSMTPMailer(cfg config.Config) (domain.Mailer, error) {
	smtp := cfg.Mail.SMTP
	opts := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(smtp.Username),
		mail.WithPassword(smtp.Password),
		mail.WithPort(smtp.Port),
	}

	switch strings.ToLower(strings.TrimSpace(smtp.TLSMode)) {
	case "tls":
		opts = append(opts, mail.WithSSLPort(true))
	case "none":
		opts = append(opts, mail.WithTLSPortPolicy(mail.NoTLS))
	default:
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(smtp.Host, opts...)
	if err != nil {
		return nil, cerrors.WithStack(err)
	}

	mailgen.SetDefault(mailgen.New().Product(mailgen.Product{
		Name: cfg.App.Name,
		Link: cfg.App.FrontendBaseURL,
	}))

	mailer := &smtpMailer{
		client:  client,
		mailCfg: cfg.Mail,
	}

	return mailer, nil
}

func (m *smtpMailer) Send(ctx context.Context, message *domain.Mail) error {
	if message == nil {
		return errors.New("mail message cannot be nil")
	}
	if len(message.To) == 0 {
		return errors.New("email recipient cannot be empty")
	}
	if message.Subject == "" {
		return errors.New("email subject cannot be empty")
	}
	msg := mail.NewMsg()
	from := fmt.Sprintf("%s <%s>", m.mailCfg.From.Name, m.mailCfg.From.Address)
	if err := msg.From(from); err != nil {
		return cerrors.WithStack(err)
	}
	if err := msg.To(message.To...); err != nil {
		return cerrors.WithStack(err)
	}
	if len(message.Cc) > 0 {
		if err := msg.Cc(message.Cc...); err != nil {
			return cerrors.WithStack(err)
		}
	}
	if len(message.Bcc) > 0 {
		if err := msg.Bcc(message.Bcc...); err != nil {
			return cerrors.WithStack(err)
		}
	}
	msg.Subject(message.Subject)
	msg.SetBodyString(mail.TypeTextPlain, message.Text)
	msg.SetBodyString(mail.TypeTextHTML, message.HTML)

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return cerrors.WithStack(err)
	}

	return nil
}

func (m *logMailer) Send(ctx context.Context, message *domain.Mail) error {
	if message == nil {
		return errors.New("mail message cannot be nil")
	}
	if len(message.To) == 0 {
		return errors.New("email recipient cannot be empty")
	}
	if message.Subject == "" {
		return errors.New("email subject cannot be empty")
	}

	slog.InfoContext(
		ctx,
		"email delivery via log driver",
		"to", message.To,
		"cc", message.Cc,
		"bcc", message.Bcc,
		"subject", message.Subject,
		"text", message.Text,
		"html", message.HTML,
	)

	return nil
}

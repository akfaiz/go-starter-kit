package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/akfaiz/go-mailgen"
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/wneessen/go-mail"
)

type smtpMailer struct {
	client  *mail.Client
	appCfg  config.App
	mailCfg config.Mail
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
		return nil, err
	}

	mailgen.SetDefault(mailgen.New().Product(mailgen.Product{
		Name: cfg.App.Name,
		Link: cfg.App.FrontendBaseURL,
	}))

	mailer := &smtpMailer{
		client:  client,
		appCfg:  cfg.App,
		mailCfg: cfg.Mail,
	}

	return mailer, nil
}

func (m *smtpMailer) Send(ctx context.Context, builder *mailgen.Builder) error {
	if builder == nil {
		return errors.New("message builder cannot be nil")
	}
	message, err := builder.Build()
	if err != nil {
		return err
	}
	if len(message.To()) == 0 {
		return errors.New("email recipient cannot be empty")
	}
	if message.Subject() == "" {
		return errors.New("email subject cannot be empty")
	}
	msg := mail.NewMsg()
	from := fmt.Sprintf("%s <%s>", m.mailCfg.From.Name, m.mailCfg.From.Address)
	if err := msg.From(from); err != nil {
		return err
	}
	if err := msg.To(message.To()...); err != nil {
		return err
	}
	if len(message.Cc()) > 0 {
		if err := msg.Cc(message.Cc()...); err != nil {
			return err
		}
	}
	if len(message.Bcc()) > 0 {
		if err := msg.Bcc(message.Bcc()...); err != nil {
			return err
		}
	}
	msg.Subject(message.Subject())
	msg.SetBodyString(mail.TypeTextPlain, message.PlainText())
	msg.SetBodyString(mail.TypeTextHTML, message.HTML())

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return err
	}

	return nil
}

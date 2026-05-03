---
name: go-email-notifications
description: Email composition with go-mailgen and SMTP delivery for notifications. Use when creating new email notifications, customizing templates, or sending emails.
---

# Email and Notifications

This project uses `github.com/akfaiz/go-mailgen` for composing notification content and `github.com/wneessen/go-mail` for SMTP delivery.

The actual transport contract is `internal/domain.Mail`, not a builder:

```go
type Mailer interface {
    Send(ctx context.Context, mail *domain.Mail) error
}

type Mail struct {
    To      []string
    Cc      []string
    Bcc     []string
    Subject string
    Text    string
    HTML    string
}
```

## Sending Flow
Notifications are composed in the service layer, serialized into a queue payload, and sent by the queue worker:

1. Service builds a `*mailgen.Builder`.
2. Service calls `Build()` to get a message with recipients, subject, plain text, and HTML.
3. Service marshals `payload.MailPayload` and enqueues `queue.TypeMailSend`.
4. `internal/delivery/queue/handler/mail_handler.go` converts the payload back to `domain.Mail`.
5. `internal/infra/smtp_client.go` sends the `domain.Mail` via SMTP.

## Creating a New Email Notification
Follow the existing forgot-password OTP pattern in `internal/service/auth/auth_service.go`.

### 1. Build the email in the service
Create a private helper that returns `*mailgen.Builder`.

```go
func (s *service) buildEmailForgotPasswordOTP(user *domain.User, otp string) *mailgen.Builder {
    return mailgen.New().
        To(user.Email).
        Subject("Password Reset OTP").
        Name(user.Name).
        Line("Use the following OTP to reset your password:").
        Action("Your OTP", otp).
        Linef("This OTP expires in %d minutes.", int(s.cfg.Auth.ResetPasswordExpiration.Minutes())).
        Line("If you did not request a password reset, ignore this email.")
}
```

### 2. Convert the builder output into a queue payload
Call `Build()`, then map the resulting message into `payload.MailPayload`.

```go
builder := s.buildEmailForgotPasswordOTP(user, otp)
message, err := builder.Build()
if err != nil {
    return err
}

payload, err := json.Marshal(payload.MailPayload{
    To:      message.To(),
    Cc:      message.Cc(),
    Bcc:     message.Bcc(),
    Subject: message.Subject(),
    Text:    message.PlainText(),
    HTML:    message.HTML(),
})
if err != nil {
    return err
}
```

### 3. Enqueue the mail task
Use `queue.TypeMailSend` with `queue.NewTask`.

```go
_, err := s.queue.EnqueueContext(ctx, queue.NewTask(ctx, queue.TypeMailSend, payload))
```

## Queue Payload
`internal/delivery/queue/handler/payload/payload.go` defines the payload shape used between service and worker.

```go
type MailPayload struct {
    To      []string `json:"to"`
    Cc      []string `json:"cc"`
    Bcc     []string `json:"bcc"`
    Subject string   `json:"subject"`
    Text    string   `json:"text"`
    HTML    string   `json:"html"`
}
```

`MailPayload.ToDomain()` returns `*domain.Mail` for the SMTP adapter.

## SMTP Delivery
`internal/infra/smtp_client.go` creates the SMTP mailer and sends `domain.Mail` messages.

### Validation and defaults
- `mail message cannot be nil`
- `email recipient cannot be empty`
- `email subject cannot be empty`
- `mailgen.SetDefault(...)` is configured with `cfg.App.Name` and `cfg.App.FrontendBaseURL`

### SMTP config
`internal/config/mail.go` reads:
- `MAIL_HOST`
- `MAIL_PORT`
- `MAIL_USERNAME`
- `MAIL_PASSWORD`
- `MAIL_TLS_MODE` with values `starttls`, `tls`, or `none`
- `MAIL_FROM_ADDRESS`
- `MAIL_FROM_NAME`

## Builder API Reference
Use the existing `mailgen.Builder` methods shown in the auth service pattern:
- `.To(emails ...string)`
- `.Subject(s string)`
- `.Name(n string)`
- `.Line(text string)`
- `.Linef(format string, args ...any)`
- `.Action(text string, link string)`
- `.Build()`

## Practical Rules
- Compose emails in the service layer, not in the SMTP adapter.
- Keep the transport boundary on `domain.Mail`.
- Preserve both plain text and HTML bodies when enqueueing mail.
- Use `payload.MailPayload` when crossing the queue boundary.

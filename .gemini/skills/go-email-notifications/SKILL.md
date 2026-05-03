---
name: go-email-notifications
description: Email templates and mailgen builder for notifications. Use when creating new email notifications, customizing templates, or sending emails.
---

# Email and Notifications

This project uses `github.com/akfaiz/go-mailgen` (a builder-pattern wrapper) to generate responsive HTML/text emails and a standard SMTP client for delivery.

## Sending Emails
The `Mailer` interface (`internal/domain/mailer.go`) uses a `*mailgen.Builder` to construct emails.

```go
type Mailer interface {
    Send(ctx context.Context, msg *mailgen.Builder) error
}
```

## Creating a New Email Notification
Emails are typically built within the service layer using a builder method.

### 1. Define the Builder Method in Service
Create a private method in your service to construct the `*mailgen.Builder`.

```go
func (s *service) buildWelcomeEmail(user *domain.User) *mailgen.Builder {
    return mailgen.New().
        To(user.Email).
        Subject("Welcome to " + s.cfg.App.Name).
        Name(user.Name).
        Line("We are excited to have you on board!").
        Action("Get Started", s.cfg.App.FrontendBaseURL + "/dashboard").
        Line("If you have any questions, just reply to this email.")
}
```

### 2. Trigger the Email
Call `mailer.Send` with the builder in your service method.

```go
func (s *service) Register(ctx context.Context, user *domain.User) error {
    // ... logic ...
    if err := s.mailer.Send(ctx, s.buildWelcomeEmail(user)); err != nil {
        return err // Or log/handle error
    }
    return nil
}
```

## Builder API Reference (`mailgen.Builder`)
- `.To(emails ...string)`: Set recipients.
- `.Subject(s string)`: Set email subject.
- `.Name(n string)`: Set the greeting name (e.g., "Hi John,").
- `.Line(text string)`: Add a paragraph of text.
- `.Linef(format string, args ...any)`: Add a formatted paragraph.
- `.Action(text string, link string)`: Add a call-to-action button.
- `.Table(table mailgen.Table)`: Add a data table.

## Configuration
SMTP and default product settings (name, link) are managed in `internal/config/` and initialized in `internal/infra/smtp_client.go`.

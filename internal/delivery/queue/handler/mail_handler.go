package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/handler/payload"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/hibiken/asynq"
)

type MailTaskHandler struct {
	mailer domain.Mailer
}

func NewMailTaskHandler(mailer domain.Mailer) *MailTaskHandler {
	return &MailTaskHandler{mailer: mailer}
}

func (h *MailTaskHandler) HandleSendMail(ctx context.Context, t *asynq.Task) error {
	var p payload.MailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w: %w", err, asynq.SkipRetry)
	}

	slog.InfoContext(ctx, "sending email via worker", "to", p.To, "subject", p.Subject)

	return h.mailer.Send(ctx, p.ToDomain())
}

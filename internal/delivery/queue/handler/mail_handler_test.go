package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/handler"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/handler/payload"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMailTaskHandler_HandleSendMail(t *testing.T) {
	t.Run("returns skip retry on invalid json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		h := handler.NewMailTaskHandler(mocks.NewMockMailer(ctrl))

		err := h.HandleSendMail(context.Background(), asynq.NewTask(queue.TypeMailSend, []byte("not-json")))
		require.Error(t, err)
		assert.ErrorIs(t, err, asynq.SkipRetry)
		assert.ErrorContains(t, err, "json.Unmarshal failed")
	})

	t.Run("sends converted mail payload", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		mailer := mocks.NewMockMailer(ctrl)
		h := handler.NewMailTaskHandler(mailer)

		want := &domain.Mail{
			To:      []string{"to@example.com"},
			Cc:      []string{"cc@example.com"},
			Bcc:     []string{"bcc@example.com"},
			Subject: "subject",
			Text:    "plain text",
			HTML:    "<p>html</p>",
		}
		body, err := json.Marshal(payload.MailPayload{
			To:      want.To,
			Cc:      want.Cc,
			Bcc:     want.Bcc,
			Subject: want.Subject,
			Text:    want.Text,
			HTML:    want.HTML,
		})
		require.NoError(t, err)

		mailer.EXPECT().
			Send(gomock.Any(), want).
			Return(nil)

		err = h.HandleSendMail(context.Background(), asynq.NewTask(queue.TypeMailSend, body))
		require.NoError(t, err)
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		mailer := mocks.NewMockMailer(ctrl)
		h := handler.NewMailTaskHandler(mailer)

		body, err := json.Marshal(payload.MailPayload{To: []string{"to@example.com"}, Subject: "subject"})
		require.NoError(t, err)

		wantErr := errors.New("send failed")
		mailer.EXPECT().
			Send(gomock.Any(), gomock.AssignableToTypeOf(&domain.Mail{})).
			Return(wantErr)

		err = h.HandleSendMail(context.Background(), asynq.NewTask(queue.TypeMailSend, body))
		require.ErrorIs(t, err, wantErr)
	})
}

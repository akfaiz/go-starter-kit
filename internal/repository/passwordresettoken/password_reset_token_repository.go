package passwordresettoken

import (
	"context"
	"database/sql"
	"errors"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/model"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("password-reset-token-repository")

type repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) domain.PasswordResetTokenRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	m := model.NewPasswordResetTokenFromDomain(token)
	_, err := r.db.NewInsert().Model(m).
		On("CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at").
		Exec(ctx)
	if err != nil {
		return err
	}
	token.ID = m.ID
	token.CreatedAt = m.CreatedAt
	return nil
}

func (r *repository) FindOne(ctx context.Context, userID int64) (*domain.PasswordResetToken, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	m := new(model.PasswordResetToken)
	err := r.db.NewSelect().Model(m).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *repository) Delete(ctx context.Context, userID int64) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	_, err := r.db.NewDelete().Model((*model.PasswordResetToken)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

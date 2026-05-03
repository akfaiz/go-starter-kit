package passwordresettoken

import (
	"context"
	"errors"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/model"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("password-reset-token-repository")

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.PasswordResetTokenRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	m := model.NewPasswordResetTokenFromDomain(token)
	err := gorm.G[model.PasswordResetToken](r.db, clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"token", "expires_at"}),
	}).Create(ctx, m)
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

	m, err := gorm.G[model.PasswordResetToken](r.db).Where("user_id = ?", userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *repository) Delete(ctx context.Context, userID int64) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	_, err := gorm.G[model.PasswordResetToken](r.db).Where("user_id = ?", userID).Delete(ctx)
	return err
}

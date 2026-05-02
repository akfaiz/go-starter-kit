package usertoken

import (
	"context"
	"database/sql"
	"errors"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/model"
	"github.com/uptrace/bun"
)

type repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) domain.UserTokenRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *domain.UserToken) error {
	m := &model.UserToken{
		UserID:    token.UserID,
		Token:     token.Token,
		TokenType: string(token.TokenType),
		ExpiresAt: token.ExpiresAt,
	}
	_, err := r.db.NewInsert().Model(m).
		On("CONFLICT (user_id, token_type) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at").
		Exec(ctx)
	if err != nil {
		return err
	}
	token.ID = m.ID
	token.CreatedAt = m.CreatedAt
	return nil
}

func (r *repository) FindOne(ctx context.Context, userID int64, tokenType domain.TokenType) (*domain.UserToken, error) {
	m := new(model.UserToken)
	err := r.db.NewSelect().Model(m).Where("user_id = ? AND token_type = ?", userID, string(tokenType)).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *repository) Delete(ctx context.Context, userID int64, tokenType domain.TokenType) error {
	_, err := r.db.NewDelete().Model((*model.UserToken)(nil)).
		Where("user_id = ? AND token_type = ?", userID, string(tokenType)).
		Exec(ctx)
	return err
}

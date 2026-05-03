package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/model"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("user-repository")

type repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) domain.UserRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	m := model.NewUserFromDomain(user)
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	if err != nil {
		var pgError pgdriver.Error
		if errors.As(err, &pgError) {
			if pgError.IntegrityViolation() && strings.Contains(pgError.Error(), "uk_users_email") {
				return domain.ErrEmailAlreadyExists
			}
		}
		return err
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return user.ToDomain(), nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user := new(model.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return user.ToDomain(), nil
}

func (r *repository) FindAll(
	ctx context.Context,
	params domain.FindAllParams,
) (*domain.Paginated[*domain.User], error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	users := make([]*model.User, 0)
	query := r.db.NewSelect().Model(&users)

	if params.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	query = applySort(query, params)

	count, err := query.
		Limit(params.Limit).
		Offset((params.Page - 1) * params.Limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, err
	}

	domainUsers := make([]*domain.User, len(users))
	for i, u := range users {
		domainUsers[i] = u.ToDomain()
	}

	return &domain.Paginated[*domain.User]{
		Items:      domainUsers,
		Pagination: domain.NewPagination(params.Page, params.Limit, int64(count)),
	}, nil
}

func applySort(query *bun.SelectQuery, params domain.FindAllParams) *bun.SelectQuery {
	if params.Sort == "" {
		return query.Order("id ASC")
	}

	allowedSortFields := map[string]string{
		"id":         "id",
		"name":       "name",
		"email":      "email",
		"created_at": "created_at",
	}

	dbField, ok := allowedSortFields[params.Sort]
	if !ok {
		return query.Order("id ASC")
	}

	order := "ASC"
	if strings.ToUpper(params.Order) == "DESC" {
		order = "DESC"
	}

	return query.Order(dbField + " " + order)
}

func (r *repository) Update(ctx context.Context, id int64, data *domain.UserUpdate) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	query := r.db.NewUpdate().Model((*model.User)(nil)).Where("id = ?", id)
	query = model.ApplyUserUpdate(query, data)
	res, err := query.Exec(ctx)
	if err != nil {
		var pgError pgdriver.Error
		if errors.As(err, &pgError) {
			if pgError.IntegrityViolation() && strings.Contains(pgError.Error(), "uk_users_email") {
				return domain.ErrEmailAlreadyExists
			}
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrResourceNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	res, err := r.db.NewDelete().Model(&model.User{ID: id}).WherePK().Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrResourceNotFound
	}
	return nil
}

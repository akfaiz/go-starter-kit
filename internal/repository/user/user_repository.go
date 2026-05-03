package user

import (
	"context"
	"errors"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/model"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

var tracer = otel.Tracer("user-repository")

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.UserRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	m := model.NewUserFromDomain(user)
	if err := gorm.G[model.User](r.db).Create(ctx, m); err != nil {
		if strings.Contains(err.Error(), "users_email_unique") {
			return domain.ErrEmailAlreadyExists
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

	user, err := gorm.G[model.User](r.db).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}
	return user.ToDomain(), nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := gorm.G[model.User](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	query := gorm.G[model.User](r.db).Where("1=1")

	if params.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	total, err := query.Count(ctx, "*")
	if err != nil {
		return nil, err
	}

	query = applySort(query, params)

	users, err := query.
		Limit(params.Limit).
		Offset((params.Page - 1) * params.Limit).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	domainUsers := make([]*domain.User, len(users))
	for i, u := range users {
		domainUsers[i] = u.ToDomain()
	}

	return &domain.Paginated[*domain.User]{
		Items:      domainUsers,
		Pagination: domain.NewPagination(params.Page, params.Limit, total),
	}, nil
}

func applySort(query gorm.ChainInterface[model.User], params domain.FindAllParams) gorm.ChainInterface[model.User] {
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

	var user model.User
	var fields []string
	if data.Name.IsValue() {
		user.Name = data.Name.MustGet()
		fields = append(fields, "name")
	}
	if data.Email.IsValue() {
		user.Email = data.Email.MustGet()
		fields = append(fields, "email")
	}
	if data.Password.IsValue() {
		user.Password = data.Password.MustGet()
		fields = append(fields, "password")
	}
	if data.EmailVerifiedAt.IsValue() {
		if data.EmailVerifiedAt.IsNull() {
			user.EmailVerifiedAt = nil
		} else {
			val := data.EmailVerifiedAt.MustGet()
			user.EmailVerifiedAt = &val
		}
		fields = append(fields, "email_verified_at")
	}

	if len(fields) == 0 {
		return nil
	}

	args := make([]any, len(fields)-1)
	for i := 1; i < len(fields); i++ {
		args[i-1] = fields[i]
	}

	rowsAffected, err := gorm.G[model.User](r.db).
		Where("id = ?", id).
		Select(fields[0], args...).
		Updates(ctx, user)

	if err != nil {
		if strings.Contains(err.Error(), "users_email_unique") {
			return domain.ErrEmailAlreadyExists
		}
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrResourceNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	rowsAffected, err := gorm.G[model.User](r.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrResourceNotFound
	}
	return nil
}

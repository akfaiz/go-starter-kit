package user

import (
	"context"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("user-service")

type service struct {
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
}

func NewService(
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
) domain.UserService {
	return &service{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

func (s *service) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	return s.userRepo.FindByID(ctx, id)
}

func (s *service) FindAll(ctx context.Context, params domain.FindAllParams) (*domain.Paginated[*domain.User], error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	return s.userRepo.FindAll(ctx, params)
}

func (s *service) UpdateProfile(ctx context.Context, id int64, user *domain.User) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	oldUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	update := &domain.UserUpdate{}
	if oldUser.Name != user.Name {
		update.Name = omit.From(user.Name)
	}
	if oldUser.Email != user.Email {
		var t *time.Time
		update.EmailVerifiedAt = omitnull.FromPtr(t)
		update.Email = omit.From(user.Email)
	}
	if update.IsEmpty() {
		return nil
	}

	return s.userRepo.Update(ctx, id, update)
}

func (s *service) ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	match, err := s.passwordHasher.Verify(currentPassword, user.Password)
	if err != nil {
		return err
	}
	if !match {
		return domain.ErrInvalidPassword
	}
	hashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}
	update := &domain.UserUpdate{
		Password: omit.From(hashedPassword),
	}
	return s.userRepo.Update(ctx, id, update)
}

func (s *service) Delete(ctx context.Context, id int64, password string) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	match, err := s.passwordHasher.Verify(password, user.Password)
	if err != nil {
		return err
	}
	if !match {
		return domain.ErrInvalidPassword
	}
	return s.userRepo.Delete(ctx, id)
}

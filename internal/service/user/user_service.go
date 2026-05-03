package user

import (
	"context"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
)

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
	return s.userRepo.FindByID(ctx, id)
}

func (s *service) UpdateProfile(ctx context.Context, id int64, user *domain.User) error {
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

	if err := s.userRepo.Update(ctx, id, update); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return validator.NewError("email", "Email already exists")
		}
		return err
	}

	return nil
}

func (s *service) ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	match, err := s.passwordHasher.Verify(currentPassword, user.Password)
	if err != nil {
		return err
	}
	if !match {
		return validator.NewError("current_password", "Current password is incorrect")
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
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	match, err := s.passwordHasher.Verify(password, user.Password)
	if err != nil {
		return err
	}
	if !match {
		return validator.NewError("password", "Password is incorrect")
	}
	return s.userRepo.Delete(ctx, id)
}

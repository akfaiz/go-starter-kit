package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/akfaiz/go-mailgen"
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("auth-service")

type service struct {
	cfg                    config.Config
	userRepo               domain.UserRepository
	passwordResetTokenRepo domain.PasswordResetTokenRepository
	sessionRepo            domain.SessionRepository
	passwordHasher         domain.PasswordHasher
	jwtManager             domain.JWTManager
	mailer                 domain.Mailer
}

func NewService(
	cfg config.Config,
	userRepo domain.UserRepository,
	passwordResetTokenRepo domain.PasswordResetTokenRepository,
	sessionRepo domain.SessionRepository,
	passwordHasher domain.PasswordHasher,
	jwtManager domain.JWTManager,
	mailer domain.Mailer,
) domain.AuthService {
	return &service{
		cfg:                    cfg,
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		sessionRepo:            sessionRepo,
		passwordHasher:         passwordHasher,
		jwtManager:             jwtManager,
		mailer:                 mailer,
	}
}

func (s *service) Register(ctx context.Context, user *domain.User) (*domain.PairToken, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	hashedPassword, err := s.passwordHasher.Hash(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.issuePairToken(ctx, &domain.JWTClaims{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}

func (s *service) Login(ctx context.Context, email, password string) (*domain.PairToken, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	match, err := s.passwordHasher.Verify(password, user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, domain.ErrInvalidCredentials
	}
	claims := &domain.JWTClaims{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	return s.issuePairToken(ctx, claims)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*domain.PairToken, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	claims, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	storedRefreshToken, err := s.sessionRepo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	if storedRefreshToken != refreshToken {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	claims = &domain.JWTClaims{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	return s.issuePairToken(ctx, claims)
}

func (s *service) SendForgotPasswordOTP(ctx context.Context, email string) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return domain.ErrUserNotFound
		}
		return err
	}

	otp, err := s.generateOTP(6)
	if err != nil {
		return err
	}
	hashedOTP, err := s.passwordHasher.Hash(otp)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(s.cfg.Auth.ResetPasswordExpiration)
	token := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     hashedOTP,
		ExpiresAt: expiresAt,
	}
	if err := s.passwordResetTokenRepo.Create(ctx, token); err != nil {
		return err
	}

	if err := s.mailer.Send(ctx, s.buildEmailForgotPasswordOTP(user, otp)); err != nil {
		return err
	}
	return nil
}

func (s *service) VerifyForgotPasswordOTP(ctx context.Context, email, otp string) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	_, err := s.validateForgotPasswordOTP(ctx, email, otp)
	return err
}

func (s *service) ResetPasswordWithOTP(ctx context.Context, email, otp, newPassword string) error {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.validateForgotPasswordOTP(ctx, email, otp)
	if err != nil {
		return err
	}

	hashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.userRepo.Update(ctx, user.ID, &domain.UserUpdate{Password: omit.From(hashedPassword)}); err != nil {
		return err
	}
	_ = s.passwordResetTokenRepo.Delete(ctx, user.ID)
	_ = s.sessionRepo.DeleteSession(ctx, user.ID)
	return nil
}

func (s *service) issuePairToken(ctx context.Context, claims *domain.JWTClaims) (*domain.PairToken, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	pairToken, err := s.jwtManager.GeneratePairToken(claims)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.SavePairToken(
		ctx,
		claims.ID,
		pairToken.AccessToken,
		pairToken.RefreshToken,
	); err != nil {
		return nil, err
	}

	return pairToken, nil
}

func (s *service) validateForgotPasswordOTP(ctx context.Context, email, otp string) (*domain.User, error) {
	ctx, span := telemetry.StartSpan(ctx, tracer)
	defer span.End()

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	stored, err := s.passwordResetTokenRepo.FindOne(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	match, err := s.passwordHasher.Verify(otp, stored.Token)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, domain.ErrInvalidToken
	}

	return user, nil
}

func (s *service) buildEmailForgotPasswordOTP(user *domain.User, otp string) *mailgen.Builder {
	return mailgen.New().
		To(user.Email).
		Subject("Password Reset OTP").
		Name(user.Name).
		Line("Use the following OTP to reset your password:").
		Action("Your OTP", otp).
		Linef("This OTP expires in %d minutes.", int(s.cfg.Auth.ResetPasswordExpiration.Minutes())).
		Line("If you did not request a password reset, ignore this email.")
}

func (s *service) generateOTP(length int) (string, error) {
	const digits = "0123456789"
	upperBound := big.NewInt(int64(len(digits)))
	otp := make([]byte, length)
	for i := range otp {
		n, err := rand.Int(rand.Reader, upperBound)
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}
	return string(otp), nil
}

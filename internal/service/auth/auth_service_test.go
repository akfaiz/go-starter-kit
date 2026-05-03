package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/lang"
	"github.com/akfaiz/go-starter-kit/internal/service/auth"
	"github.com/akfaiz/go-starter-kit/test/mocks"
	"github.com/invopop/ctxi18n"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

var _ = BeforeSuite(func() {
	lang.Init()
})

var _ = Describe("Auth", Label("unit", "usecase"), func() {
	var (
		userRepoMock               *mocks.MockUserRepository
		passwordResetTokenRepoMock *mocks.MockPasswordResetTokenRepository
		sessionRepoMock            *mocks.MockSessionRepository
		hasherMock                 *mocks.MockPasswordHasher
		jwtManagerMock             *mocks.MockJWTManager
		mailerMock                 *mocks.MockMailer
		cfg                        config.Config
		svc                        domain.AuthService

		ctx context.Context
	)
	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		userRepoMock = mocks.NewMockUserRepository(ctrl)
		passwordResetTokenRepoMock = mocks.NewMockPasswordResetTokenRepository(ctrl)
		sessionRepoMock = mocks.NewMockSessionRepository(ctrl)
		hasherMock = mocks.NewMockPasswordHasher(ctrl)
		jwtManagerMock = mocks.NewMockJWTManager(ctrl)
		mailerMock = mocks.NewMockMailer(ctrl)
		cfg = config.Config{
			Auth: config.Auth{
				JWT: config.JWT{
					AccessExpires:  15 * time.Minute,
					RefreshExpires: 24 * time.Hour,
				},
			},
		}
		svc = auth.NewService(
			cfg,
			userRepoMock,
			passwordResetTokenRepoMock,
			sessionRepoMock,
			hasherMock,
			jwtManagerMock,
			mailerMock,
		)

		ctx = context.Background()
		ctx, _ = ctxi18n.WithLocale(ctx, "en")

		DeferCleanup(func() {
			ctrl.Finish()
		})
	})

	Describe("Login", func() {
		type args struct {
			email    string
			password string
		}
		type testCase struct {
			args    args
			arrange func()
			check   func(token *domain.PairToken, err error)
		}
		DescribeTable("Login scenarios",
			func(tc testCase) {
				if tc.arrange != nil {
					tc.arrange()
				}
				token, err := svc.Login(ctx, tc.args.email, tc.args.password)
				tc.check(token, err)
			},
			Entry("should return token when email and password match", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "password123",
				},
				arrange: func() {
					user := &domain.User{
						ID:    1,
						Name:  "John Doe",
						Email: "john.doe@example.com",
					}
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(user, nil)
					hasherMock.EXPECT().Verify("password123", user.Password).Return(true, nil)
					token := &domain.PairToken{
						AccessToken:  "access.token.here",
						RefreshToken: "refresh.token.here",
					}
					jwtManagerMock.EXPECT().GeneratePairToken(&domain.JWTClaims{
						ID:    user.ID,
						Name:  user.Name,
						Email: user.Email,
					}).Return(token, nil)
					sessionRepoMock.EXPECT().SavePairToken(
						gomock.Any(),
						user.ID,
						"access.token.here",
						"refresh.token.here",
					).Return(nil)
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).NotTo(HaveOccurred())
					Expect(token).NotTo(BeNil())
					Expect(token.AccessToken).To(Equal("access.token.here"))
					Expect(token.RefreshToken).To(Equal("refresh.token.here"))
				},
			}),
			Entry("should return error when email not found", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "password123",
				},
				arrange: func() {
					userRepoMock.EXPECT().
						FindByEmail(gomock.Any(), "john.doe@example.com").
						Return(nil, domain.ErrResourceNotFound)
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					Expect(errors.Is(err, domain.ErrInvalidCredentials)).To(BeTrue())
				},
			}),
			Entry("should return error when password verification fails", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "wrongpassword",
				},
				arrange: func() {
					user := &domain.User{
						ID:    1,
						Name:  "John Doe",
						Email: "john.doe@example.com",
					}
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(user, nil)
					hasherMock.EXPECT().Verify("wrongpassword", user.Password).Return(false, nil)
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					Expect(errors.Is(err, domain.ErrInvalidCredentials)).To(BeTrue())
				},
			}),
			Entry("should return error when password hasher fails", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "password123",
				},
				arrange: func() {
					user := &domain.User{
						ID:    1,
						Name:  "John Doe",
						Email: "john.doe@example.com",
					}
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(user, nil)
					hasherMock.EXPECT().Verify("password123", user.Password).Return(false, errors.New("hash error"))
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("hash error"))
				},
			}),
			Entry("should return error when JWT generation fails", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "password123",
				},
				arrange: func() {
					user := &domain.User{
						ID:    1,
						Name:  "John Doe",
						Email: "john.doe@example.com",
					}
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(user, nil)
					hasherMock.EXPECT().Verify("password123", user.Password).Return(true, nil)
					jwtManagerMock.EXPECT().
						GeneratePairToken(gomock.Any()).
						Return(nil, errors.New("jwt generation failed"))
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("jwt generation failed"))
				},
			}),
			Entry("should return error when session save fails", testCase{
				args: args{
					email:    "john.doe@example.com",
					password: "password123",
				},
				arrange: func() {
					user := &domain.User{
						ID:    1,
						Name:  "John Doe",
						Email: "john.doe@example.com",
					}
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(user, nil)
					hasherMock.EXPECT().Verify("password123", user.Password).Return(true, nil)
					token := &domain.PairToken{
						AccessToken:  "access.token.here",
						RefreshToken: "refresh.token.here",
					}
					jwtManagerMock.EXPECT().GeneratePairToken(gomock.Any()).Return(token, nil)
					sessionRepoMock.EXPECT().SavePairToken(
						gomock.Any(),
						user.ID,
						"access.token.here",
						"refresh.token.here",
					).Return(errors.New("session save failed"))
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("session save failed"))
				},
			}),
		)
	})

	Describe("RefreshToken", func() {
		It("should rotate refresh token successfully", func() {
			user := &domain.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: user.ID}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), user.ID).Return("old.refresh.token", nil)
			userRepoMock.EXPECT().FindByID(gomock.Any(), user.ID).Return(user, nil)
			jwtManagerMock.EXPECT().
				GeneratePairToken(gomock.AssignableToTypeOf(&domain.JWTClaims{})).
				Return(&domain.PairToken{
					AccessToken:  "new.access.token",
					RefreshToken: "new.refresh.token",
				}, nil)
			sessionRepoMock.EXPECT().SavePairToken(
				gomock.Any(),
				user.ID,
				"new.access.token",
				"new.refresh.token",
			).Return(nil)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeNil())
			Expect(token.AccessToken).To(Equal("new.access.token"))
			Expect(token.RefreshToken).To(Equal("new.refresh.token"))
		})

		It("should return unauthorized when stored token is missing", func() {
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: 1}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), int64(1)).Return("", domain.ErrResourceNotFound)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
			Expect(errors.Is(err, domain.ErrInvalidToken)).To(BeTrue())
		})

		It("should return error when JWT verification fails", func() {
			jwtManagerMock.EXPECT().VerifyRefreshToken("invalid.token").Return(nil, errors.New("invalid token"))

			token, err := svc.RefreshToken(ctx, "invalid.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
			Expect(errors.Is(err, domain.ErrInvalidToken)).To(BeTrue())
		})

		It("should return unauthorized when token mismatch", func() {
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: 1}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), int64(1)).Return("different.token", nil)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
			Expect(errors.Is(err, domain.ErrInvalidToken)).To(BeTrue())
		})

		It("should return error when user not found", func() {
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: 1}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), int64(1)).Return("old.refresh.token", nil)
			userRepoMock.EXPECT().FindByID(gomock.Any(), int64(1)).Return(nil, domain.ErrResourceNotFound)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
		})

		It("should return error when JWT generation fails", func() {
			user := &domain.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: user.ID}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), user.ID).Return("old.refresh.token", nil)
			userRepoMock.EXPECT().FindByID(gomock.Any(), user.ID).Return(user, nil)
			jwtManagerMock.EXPECT().
				GeneratePairToken(gomock.AssignableToTypeOf(&domain.JWTClaims{})).
				Return(nil, errors.New("jwt generation failed"))

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
		})

		It("should return error when session save fails", func() {
			user := &domain.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: user.ID}, nil)
			sessionRepoMock.EXPECT().GetRefreshToken(gomock.Any(), user.ID).Return("old.refresh.token", nil)
			userRepoMock.EXPECT().FindByID(gomock.Any(), user.ID).Return(user, nil)
			jwtManagerMock.EXPECT().
				GeneratePairToken(gomock.AssignableToTypeOf(&domain.JWTClaims{})).
				Return(&domain.PairToken{
					AccessToken:  "new.access.token",
					RefreshToken: "new.refresh.token",
				}, nil)
			sessionRepoMock.EXPECT().SavePairToken(
				gomock.Any(),
				user.ID,
				"new.access.token",
				"new.refresh.token",
			).Return(errors.New("session save failed"))

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
		})
	})

	Describe("Register", func() {
		It("should register user and return token successfully", func() {
			newUser := &domain.User{
				Name:     "Jane Doe",
				Email:    "jane@example.com",
				Password: "password123",
			}
			hasherMock.EXPECT().Hash("password123").Return("hashed-password", nil)
			userRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, u *domain.User) {
				u.ID = 2
				u.CreatedAt = time.Now()
				u.UpdatedAt = time.Now()
			}).Return(nil)
			jwtManagerMock.EXPECT().
				GeneratePairToken(gomock.AssignableToTypeOf(&domain.JWTClaims{})).
				Return(&domain.PairToken{
					AccessToken:  "access.token",
					RefreshToken: "refresh.token",
				}, nil)
			sessionRepoMock.EXPECT().SavePairToken(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(nil)

			token, err := svc.Register(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeNil())
			Expect(newUser.ID).To(Equal(int64(2)))
		})

		It("should return error when email already exists", func() {
			newUser := &domain.User{
				Name:     "Jane Doe",
				Email:    "existing@example.com",
				Password: "password123",
			}
			hasherMock.EXPECT().Hash("password123").Return("hashed-password", nil)
			userRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.ErrEmailAlreadyExists)

			token, err := svc.Register(ctx, newUser)
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
			Expect(errors.Is(err, domain.ErrEmailAlreadyExists)).To(BeTrue())
		})

		It("should return error when password hashing fails", func() {
			newUser := &domain.User{
				Name:     "Jane Doe",
				Email:    "jane@example.com",
				Password: "password123",
			}
			hasherMock.EXPECT().Hash("password123").Return("", errors.New("hash failed"))

			token, err := svc.Register(ctx, newUser)
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())
		})
	})

	Describe("SendForgotPasswordOTP", func() {
		It("should send OTP email successfully", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			hasherMock.EXPECT().Hash(gomock.Any()).Return("hashed-otp", nil)
			passwordResetTokenRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			mailerMock.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

			err := svc.SendForgotPasswordOTP(ctx, "john@example.com")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userRepoMock.EXPECT().
				FindByEmail(gomock.Any(), "missing@example.com").
				Return(nil, domain.ErrResourceNotFound)

			err := svc.SendForgotPasswordOTP(ctx, "missing@example.com")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrUserNotFound)).To(BeTrue())
		})

		It("should return error when OTP hashing fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			hasherMock.EXPECT().Hash(gomock.Any()).Return("", errors.New("hash failed"))

			err := svc.SendForgotPasswordOTP(ctx, "john@example.com")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when token creation fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			hasherMock.EXPECT().Hash(gomock.Any()).Return("hashed-otp", nil)
			passwordResetTokenRepoMock.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(errors.New("token creation failed"))

			err := svc.SendForgotPasswordOTP(ctx, "john@example.com")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when email sending fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			hasherMock.EXPECT().Hash(gomock.Any()).Return("hashed-otp", nil)
			passwordResetTokenRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			mailerMock.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("email send failed"))

			err := svc.SendForgotPasswordOTP(ctx, "john@example.com")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("VerifyForgotPasswordOTP", func() {
		It("should verify OTP successfully", func() {
			user := &domain.User{ID: 1, Email: "john@example.com", Password: "hashed-password"}
			otp := "123456"
			hashedOTP := "hashed-123456"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}, nil)
			hasherMock.EXPECT().Verify(otp, hashedOTP).Return(true, nil)

			err := svc.VerifyForgotPasswordOTP(ctx, "john@example.com", otp)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userRepoMock.EXPECT().
				FindByEmail(gomock.Any(), "missing@example.com").
				Return(nil, domain.ErrResourceNotFound)

			err := svc.VerifyForgotPasswordOTP(ctx, "missing@example.com", "123456")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrUserNotFound)).To(BeTrue())
		})

		It("should return error when token not found", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(nil, domain.ErrResourceNotFound)

			err := svc.VerifyForgotPasswordOTP(ctx, "john@example.com", "123456")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrInvalidToken)).To(BeTrue())
		})

		It("should return error when OTP is expired", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			hashedOTP := "hashed-123456"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(-10 * time.Minute), // Expired
			}, nil)

			err := svc.VerifyForgotPasswordOTP(ctx, "john@example.com", "123456")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrTokenExpired)).To(BeTrue())
		})

		It("should return error when OTP verification fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			otp := "123456"
			hashedOTP := "hashed-654321"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}, nil)
			hasherMock.EXPECT().Verify(otp, hashedOTP).Return(false, nil)

			err := svc.VerifyForgotPasswordOTP(ctx, "john@example.com", otp)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrInvalidToken)).To(BeTrue())
		})
	})

	Describe("ResetPasswordWithOTP", func() {
		It("should reset password successfully", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			otp := "123456"
			hashedOTP := "hashed-123456"
			newPassword := "newpassword123"
			hashedPassword := "hashed-newpassword"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}, nil)
			hasherMock.EXPECT().Verify(otp, hashedOTP).Return(true, nil)
			hasherMock.EXPECT().Hash(newPassword).Return(hashedPassword, nil)
			userRepoMock.EXPECT().Update(gomock.Any(), user.ID, gomock.Any()).Return(nil)
			passwordResetTokenRepoMock.EXPECT().Delete(gomock.Any(), user.ID).Return(nil)
			sessionRepoMock.EXPECT().DeleteSession(gomock.Any(), user.ID).Return(nil)

			err := svc.ResetPasswordWithOTP(ctx, "john@example.com", otp, newPassword)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userRepoMock.EXPECT().
				FindByEmail(gomock.Any(), "missing@example.com").
				Return(nil, domain.ErrResourceNotFound)

			err := svc.ResetPasswordWithOTP(ctx, "missing@example.com", "123456", "newpass")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, domain.ErrUserNotFound)).To(BeTrue())
		})

		It("should return error when password hashing fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			otp := "123456"
			hashedOTP := "hashed-123456"
			newPassword := "newpassword123"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}, nil)
			hasherMock.EXPECT().Verify(otp, hashedOTP).Return(true, nil)
			hasherMock.EXPECT().Hash(newPassword).Return("", errors.New("hash failed"))

			err := svc.ResetPasswordWithOTP(ctx, "john@example.com", otp, newPassword)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when user update fails", func() {
			user := &domain.User{ID: 1, Email: "john@example.com"}
			otp := "123456"
			hashedOTP := "hashed-123456"
			newPassword := "newpassword123"
			hashedPassword := "hashed-newpassword"

			userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john@example.com").Return(user, nil)
			passwordResetTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID).Return(&domain.PasswordResetToken{
				UserID:    user.ID,
				Token:     hashedOTP,
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}, nil)
			hasherMock.EXPECT().Verify(otp, hashedOTP).Return(true, nil)
			hasherMock.EXPECT().Hash(newPassword).Return(hashedPassword, nil)
			userRepoMock.EXPECT().Update(gomock.Any(), user.ID, gomock.Any()).Return(errors.New("update failed"))

			err := svc.ResetPasswordWithOTP(ctx, "john@example.com", otp, newPassword)
			Expect(err).To(HaveOccurred())
		})
	})
})

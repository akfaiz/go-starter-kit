package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/errdefs"
	"github.com/invopop/ctxi18n"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/mocks"
	"github.com/akfaiz/go-starter-kit/internal/service/auth"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
)

var _ = Describe("Auth", Label("unit", "usecase"), func() {
	var (
		userRepoMock      *mocks.MockUserRepository
		userTokenRepoMock *mocks.MockUserTokenRepository
		hasherMock        *mocks.MockPasswordHasher
		jwtManagerMock    *mocks.MockJWTManager
		mailerMock        *mocks.MockMailer
		cfg               config.Config
		svc               domain.AuthService

		ctx context.Context
	)
	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		userRepoMock = mocks.NewMockUserRepository(ctrl)
		userTokenRepoMock = mocks.NewMockUserTokenRepository(ctrl)
		hasherMock = mocks.NewMockPasswordHasher(ctrl)
		jwtManagerMock = mocks.NewMockJWTManager(ctrl)
		mailerMock = mocks.NewMockMailer(ctrl)
		cfg = config.Config{
			Auth: config.Auth{
				JWT: config.JWT{
					RefreshExpires: 24 * time.Hour,
				},
			},
		}
		svc = auth.NewService(cfg, userRepoMock, userTokenRepoMock, hasherMock, jwtManagerMock, mailerMock)

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
					hasherMock.EXPECT().Hash("refresh.token.here").Return("hashed-refresh-token", nil)
					userTokenRepoMock.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserToken{})).DoAndReturn(
						func(_ context.Context, t *domain.UserToken) error {
							Expect(t.UserID).To(Equal(user.ID))
							Expect(t.TokenType).To(Equal(domain.TokenTypeRefreshToken))
							Expect(t.Token).To(Equal("hashed-refresh-token"))
							return nil
						},
					)
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
					userRepoMock.EXPECT().FindByEmail(gomock.Any(), "john.doe@example.com").Return(nil, domain.ErrResourceNotFound)
				},
				check: func(token *domain.PairToken, err error) {
					Expect(err).To(HaveOccurred())
					Expect(token).To(BeNil())
					var vErr *validator.ValidationError
					Expect(errors.As(err, &vErr)).To(BeTrue())
					Expect(vErr.First().Field).To(Equal("email"))
					Expect(vErr.First().Message).To(Equal("These credentials do not match our records."))
				},
			}),
		)
	})

	Describe("RefreshToken", func() {
		It("should rotate refresh token successfully", func() {
			user := &domain.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: user.ID}, nil)
			userTokenRepoMock.EXPECT().FindOne(gomock.Any(), user.ID, domain.TokenTypeRefreshToken).Return(&domain.UserToken{
				UserID:    user.ID,
				Token:     "old.refresh.hash",
				TokenType: domain.TokenTypeRefreshToken,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil)
			hasherMock.EXPECT().Verify("old.refresh.token", "old.refresh.hash").Return(true, nil)
			userRepoMock.EXPECT().FindByID(gomock.Any(), user.ID).Return(user, nil)
			jwtManagerMock.EXPECT().GeneratePairToken(gomock.AssignableToTypeOf(&domain.JWTClaims{})).Return(&domain.PairToken{
				AccessToken:  "new.access.token",
				RefreshToken: "new.refresh.token",
			}, nil)
			hasherMock.EXPECT().Hash("new.refresh.token").Return("new.refresh.hash", nil)
			userTokenRepoMock.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserToken{})).DoAndReturn(
				func(_ context.Context, t *domain.UserToken) error {
					Expect(t.UserID).To(Equal(user.ID))
					Expect(t.TokenType).To(Equal(domain.TokenTypeRefreshToken))
					Expect(t.Token).To(Equal("new.refresh.hash"))
					return nil
				},
			)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeNil())
			Expect(token.AccessToken).To(Equal("new.access.token"))
			Expect(token.RefreshToken).To(Equal("new.refresh.token"))
		})

		It("should return unauthorized when stored token is missing", func() {
			jwtManagerMock.EXPECT().VerifyRefreshToken("old.refresh.token").Return(&domain.JWTClaims{ID: 1}, nil)
			userTokenRepoMock.EXPECT().FindOne(gomock.Any(), int64(1), domain.TokenTypeRefreshToken).Return(nil, domain.ErrResourceNotFound)

			token, err := svc.RefreshToken(ctx, "old.refresh.token")
			Expect(err).To(HaveOccurred())
			Expect(token).To(BeNil())

			var appErr *errdefs.AppError
			Expect(errors.As(err, &appErr)).To(BeTrue())
			Expect(appErr.Status).To(Equal(401))
		})
	})
})

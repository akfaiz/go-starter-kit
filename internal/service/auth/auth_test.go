package auth_test

import (
	"context"
	"errors"

	"github.com/invopop/ctxi18n"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/mocks"
	"github.com/akfaiz/go-starter-kit/internal/service/auth"
	"github.com/akfaiz/go-starter-kit/internal/validator"
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

		ctx    context.Context
	)
	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		userRepoMock = mocks.NewMockUserRepository(ctrl)
		userTokenRepoMock = mocks.NewMockUserTokenRepository(ctrl)
		hasherMock = mocks.NewMockPasswordHasher(ctrl)
		jwtManagerMock = mocks.NewMockJWTManager(ctrl)
		mailerMock = mocks.NewMockMailer(ctrl)
		cfg = config.Config{}
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
})

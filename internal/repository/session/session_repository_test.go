package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/repository/session"
	"github.com/alicebob/miniredis/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

func TestSessionRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Session Repository Suite")
}

var _ = Describe("Session Repository", Label("unit", "repository"), func() {
	var (
		mr     *miniredis.Miniredis
		client *redis.Client
		r      domain.SessionRepository
		ctx    context.Context
		cfg    config.Config
	)

	BeforeEach(func() {
		mr = miniredis.NewMiniRedis()
		Expect(mr.Start()).To(BeNil())

		client = redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		cfg = config.Config{
			Redis: config.Redis{
				Prefix: "test",
			},
			Auth: config.Auth{
				JWT: config.JWT{
					AccessExpires:  15 * time.Minute,
					RefreshExpires: 7 * 24 * time.Hour,
				},
			},
		}

		r = session.NewRepository(cfg, client)
		ctx = context.Background()

		DeferCleanup(func() {
			_ = client.Close()
			mr.Close()
		})
	})

	Describe("SavePairToken", func() {
		var (
			userID       int64
			accessToken  string
			refreshToken string
			err          error
		)

		BeforeEach(func() {
			userID = 1
			accessToken = "access-token-123"
			refreshToken = "refresh-token-456"
		})

		JustBeforeEach(func() {
			err = r.SavePairToken(ctx, userID, accessToken, refreshToken)
		})

		When("saving tokens successfully", func() {
			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should store access token in Redis", func() {
				token, err := client.Get(ctx, "test:auth:session:1:access").Result()
				Expect(err).To(BeNil())
				Expect(token).To(Equal(accessToken))
			})

			It("should store refresh token in Redis", func() {
				token, err := client.Get(ctx, "test:auth:session:1:refresh").Result()
				Expect(err).To(BeNil())
				Expect(token).To(Equal(refreshToken))
			})
		})

		When("saving tokens for different users", func() {
			It("should store tokens separately", func() {
				err := r.SavePairToken(ctx, 2, "access-2", "refresh-2")
				Expect(err).To(BeNil())

				token1, _ := client.Get(ctx, "test:auth:session:1:access").Result()
				token2, _ := client.Get(ctx, "test:auth:session:2:access").Result()

				Expect(token1).To(Equal(accessToken))
				Expect(token2).To(Equal("access-2"))
			})
		})

		When("overwriting existing tokens", func() {
			BeforeEach(func() {
				// Save initial tokens
				err = r.SavePairToken(ctx, userID, "old-access", "old-refresh")
				Expect(err).To(BeNil())

				// Update with new tokens
				accessToken = "new-access"
				refreshToken = "new-refresh"
			})

			It("should replace old tokens with new ones", func() {
				access, _ := r.GetAccessToken(ctx, userID)
				refresh, _ := r.GetRefreshToken(ctx, userID)

				Expect(access).To(Equal("new-access"))
				Expect(refresh).To(Equal("new-refresh"))
			})
		})
	})

	Describe("GetAccessToken", func() {
		var (
			userID int64
			token  string
			err    error
			actErr error
		)

		BeforeEach(func() {
			userID = 1
			token = "access-token-123"
		})

		When("token exists", func() {
			BeforeEach(func() {
				err = r.SavePairToken(ctx, userID, token, "refresh-token")
				Expect(err).To(BeNil())
			})

			JustBeforeEach(func() {
				token, actErr = r.GetAccessToken(ctx, userID)
			})

			It("should return the access token", func() {
				Expect(actErr).To(BeNil())
				Expect(token).To(Equal("access-token-123"))
			})
		})

		When("token does not exist", func() {
			JustBeforeEach(func() {
				token, actErr = r.GetAccessToken(ctx, 999)
			})

			It("should return resource not found error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return empty token", func() {
				Expect(token).To(Equal(""))
			})
		})
	})

	Describe("GetRefreshToken", func() {
		var (
			userID int64
			token  string
			err    error
			actErr error
		)

		BeforeEach(func() {
			userID = 1
			token = "refresh-token-456"
		})

		When("token exists", func() {
			BeforeEach(func() {
				err = r.SavePairToken(ctx, userID, "access-token", token)
				Expect(err).To(BeNil())
			})

			JustBeforeEach(func() {
				token, actErr = r.GetRefreshToken(ctx, userID)
			})

			It("should return the refresh token", func() {
				Expect(actErr).To(BeNil())
				Expect(token).To(Equal("refresh-token-456"))
			})
		})

		When("token does not exist", func() {
			JustBeforeEach(func() {
				token, actErr = r.GetRefreshToken(ctx, 999)
			})

			It("should return resource not found error", func() {
				Expect(actErr).To(Equal(domain.ErrResourceNotFound))
			})

			It("should return empty token", func() {
				Expect(token).To(Equal(""))
			})
		})
	})

	Describe("DeleteSession", func() {
		var (
			userID int64
			err    error
		)

		BeforeEach(func() {
			userID = 1
		})

		JustBeforeEach(func() {
			err = r.DeleteSession(ctx, userID)
		})

		When("deleting existing session", func() {
			BeforeEach(func() {
				err = r.SavePairToken(ctx, userID, "access-token", "refresh-token")
				Expect(err).To(BeNil())
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})

			It("should delete access token", func() {
				_, err := r.GetAccessToken(ctx, userID)
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})

			It("should delete refresh token", func() {
				_, err := r.GetRefreshToken(ctx, userID)
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})

		When("deleting non-existent session", func() {
			BeforeEach(func() {
				userID = 999
			})

			It("should not return error", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("TTL expiration", func() {
		Describe("AccessToken TTL", func() {
			It("should expire after configured access TTL", func() {
				userID := int64(1)
				err := r.SavePairToken(ctx, userID, "access-token", "refresh-token")
				Expect(err).To(BeNil())

				ttl := client.TTL(ctx, "test:auth:session:1:access").Val()
				Expect(ttl).To(BeNumerically(">", 0))
				Expect(ttl).To(BeNumerically("<=", 15*time.Minute))

				// Fast forward beyond access token TTL
				mr.FastForward(16 * time.Minute)

				// Verify access token is expired
				_, err = r.GetAccessToken(ctx, userID)
				Expect(err).To(Equal(domain.ErrResourceNotFound))
			})
		})

		Describe("RefreshToken TTL", func() {
			It("should persist longer than access token TTL", func() {
				userID := int64(1)
				err := r.SavePairToken(ctx, userID, "access-token", "refresh-token")
				Expect(err).To(BeNil())

				ttl := client.TTL(ctx, "test:auth:session:1:refresh").Val()
				Expect(ttl).To(BeNumerically(">", 0))
				Expect(ttl).To(BeNumerically("<=", 7*24*time.Hour))

				// Fast forward beyond access token TTL but within refresh TTL
				mr.FastForward(16 * time.Minute)

				// Verify refresh token still exists
				token, err := r.GetRefreshToken(ctx, userID)
				Expect(err).To(BeNil())
				Expect(token).To(Equal("refresh-token"))
			})
		})
	})
})

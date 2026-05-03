package domain_test

import (
	"testing"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUser_IsVerified(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		verified *time.Time
		expected bool
	}{
		{
			name:     "verified is not nil",
			verified: &now,
			expected: true,
		},
		{
			name:     "verified is nil",
			verified: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &domain.User{
				EmailVerifiedAt: tt.verified,
			}
			assert.Equal(t, tt.expected, u.IsVerified())
		})
	}
}

func TestUserUpdate_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		update   domain.UserUpdate
		expected bool
	}{
		{
			name:     "all fields unset",
			update:   domain.UserUpdate{},
			expected: true,
		},
		{
			name: "name set",
			update: domain.UserUpdate{
				Name: omit.From("John"),
			},
			expected: false,
		},
		{
			name: "email set",
			update: domain.UserUpdate{
				Email: omit.From("john@example.com"),
			},
			expected: false,
		},
		{
			name: "password set",
			update: domain.UserUpdate{
				Password: omit.From("secret"),
			},
			expected: false,
		},
		{
			name: "email_verified_at set",
			update: domain.UserUpdate{
				EmailVerifiedAt: omitnull.From(time.Now()),
			},
			expected: false,
		},
		{
			name: "email_verified_at set to null",
			update: domain.UserUpdate{
				EmailVerifiedAt: omitnull.FromPtr((*time.Time)(nil)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.update.IsEmpty())
		})
	}
}

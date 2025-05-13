package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		isRefresh bool
		ttl       time.Duration
	}{
		{
			name:      "access token",
			userID:    "user123",
			isRefresh: false,
			ttl:       15 * time.Minute,
		},
		{
			name:      "refresh token",
			userID:    "userABC",
			isRefresh: true,
			ttl:       24 * time.Hour,
		},
	}

	secret := "test-secret"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStr, err := GenerateJWT(tt.userID, secret, tt.isRefresh, tt.ttl)
			if err != nil {
				t.Fatalf("GenerateJWT() error = %v", err)
			}
			if tokenStr == "" {
				t.Fatal("GenerateJWT() returned empty token")
			}
			// Basic format check: JWT has three parts
			parts := strings.Split(tokenStr, ".")
			if len(parts) != 3 {
				t.Fatalf("GenerateJWT() returned invalid JWT format: %s", tokenStr)
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	secret := "another-secret"
	validTTL := 1 * time.Hour
	// create a valid access and refresh token
	accessToken, err := GenerateJWT("u1", secret, false, validTTL)
	if err != nil {
		t.Fatalf("setup: GenerateJWT access failed: %v", err)
	}
	refreshToken, err := GenerateJWT("u1", secret, true, validTTL)
	if err != nil {
		t.Fatalf("setup: GenerateJWT refresh failed: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		secretKey   string
		wantUser    string
		wantRefresh bool
		wantErr     bool
	}{
		{
			name:        "valid access token",
			token:       accessToken,
			secretKey:   secret,
			wantUser:    "u1",
			wantRefresh: false,
			wantErr:     false,
		},
		{
			name:        "valid refresh token",
			token:       refreshToken,
			secretKey:   secret,
			wantUser:    "u1",
			wantRefresh: true,
			wantErr:     false,
		},
		{
			name:      "expired token",
			token:     func() string { tok, _ := GenerateJWT("u1", secret, false, -time.Minute); return tok }(),
			secretKey: secret,
			wantErr:   true,
		},
		{
			name:      "wrong secret",
			token:     accessToken,
			secretKey: "bad-secret",
			wantErr:   true,
		},
		{
			name:      "not a token",
			token:     "not.a.jwt",
			secretKey: secret,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, isRef, _, err := ParseJWT(tt.token, tt.secretKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
				assert.Equal(t, tt.wantRefresh, isRef)
			}
			if err != nil {
				return
			}
		})
	}
}

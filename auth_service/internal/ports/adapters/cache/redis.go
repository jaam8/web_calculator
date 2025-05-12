package cache

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"time"
)

type AuthCacheAdapter struct {
	client            *redis.Client
	RefreshExpiration time.Duration
	AccessExpiration  time.Duration
}

func refreshKey(token string) string {
	return fmt.Sprintf("refresh:%s", token)
}
func accessKey(token string) string {
	return fmt.Sprintf("access:%s", token)
}

func NewAuthCacheAdapter(client *redis.Client, refreshExpiration, accessExpiration time.Duration) *AuthCacheAdapter {
	return &AuthCacheAdapter{
		client:            client,
		RefreshExpiration: refreshExpiration,
		AccessExpiration:  accessExpiration,
	}
}

func (a *AuthCacheAdapter) SaveToken(token, userID string, refresh bool) error {
	switch refresh {
	case true:
		result := a.client.Set(refreshKey(token), userID, a.RefreshExpiration)
		if result.Err() != nil {
			return fmt.Errorf("failed to save refresh token: %w", result.Err())
		}
	case false:
		result := a.client.Set(accessKey(token), userID, a.AccessExpiration)
		if result.Err() != nil {
			return fmt.Errorf("failed to save access token: %w", result.Err())
		}
	}
	return nil
}

func (a *AuthCacheAdapter) GetToken(token string, refresh bool) (string, error) {
	result := &redis.StringCmd{}
	var err error

	switch refresh {
	case true:
		result = a.client.Get(refreshKey(token))
		err = fmt.Errorf("failed to get refresh token: %w", result.Err())
	case false:
		result = a.client.Get(accessKey(token))
		err = fmt.Errorf("failed to get access token: %w", result.Err())
	}
	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			return "", errs.ErrTokenExpired
		}
		return "", err
	}
	return result.Val(), nil
}

func (a *AuthCacheAdapter) DeleteToken(token string, refresh bool) error {
	switch refresh {
	case true:
		result := a.client.Del(refreshKey(token))
		if result.Err() != nil {
			return fmt.Errorf("failed to delete refresh token: %w", result.Err())
		}
	case false:
		result := a.client.Del(accessKey(token))
		if result.Err() != nil {
			return fmt.Errorf("failed to delete access token: %w", result.Err())
		}
	}
	return nil
}

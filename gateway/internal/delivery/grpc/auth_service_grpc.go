package grpc

import (
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/callers"
	auth "github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/gateway/internal/ports"
	"time"
)

type AuthService struct {
	authAdapter *ports.AuthServiceAdapter
	MaxRetries  uint
	BaseDelay   time.Duration
}

func NewAuthService(authAdapter ports.AuthServiceAdapter,
	maxRetries uint, baseDelay time.Duration) *AuthService {
	return &AuthService{
		authAdapter: &authAdapter,
		MaxRetries:  maxRetries,
		BaseDelay:   baseDelay,
	}
}

func (s *AuthService) Login(request *auth.LoginRequest) (*auth.LoginResponse, error) {
	resultChan := make(chan *auth.LoginResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.authAdapter).Login(request)
		if err != nil {
			return fmt.Errorf("error in retry Login caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call Login: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

func (s *AuthService) Register(request *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	resultChan := make(chan *auth.RegisterResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.authAdapter).Register(request)
		if err != nil {
			return fmt.Errorf("error in retry Register caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call Register: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

func (s *AuthService) Refresh(request *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	resultChan := make(chan *auth.RefreshResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.authAdapter).Refresh(request)
		if err != nil {
			return fmt.Errorf("error in retry Refresh caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call Refresh: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

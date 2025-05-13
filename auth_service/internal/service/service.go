package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jaam8/web_calculator/auth_service/internal/ports"
	"github.com/jaam8/web_calculator/auth_service/internal/service/utils"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"go.uber.org/zap"
	"time"
)

type AuthService struct {
	auth_service.AuthServiceServer
	storage           ports.StorageAdapter
	cache             ports.CacheAdapter
	jwtSecret         string
	RefreshExpiration time.Duration
	AccessExpiration  time.Duration
}

func NewAuthService(storage ports.StorageAdapter, cache ports.CacheAdapter,
	jwtSecret string, refreshExpiration, accessExpiration time.Duration) *AuthService {
	return &AuthService{
		storage:           storage,
		cache:             cache,
		jwtSecret:         jwtSecret,
		RefreshExpiration: refreshExpiration,
		AccessExpiration:  accessExpiration,
	}
}

func (s *AuthService) Register(ctx context.Context, req *auth_service.RegisterRequest) (*auth_service.RegisterResponse, error) {
	if req.GetLogin() == "" {
		return nil, errs.ErrEmptyLogin
	}
	if req.GetPassword() == "" {
		return nil, errs.ErrEmptyPassword
	}

	hashPassword, err := utils.GenerateHash(req.Password)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to generate hash password",
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate hash password: %w", err)
	}

	userID, err := s.storage.RegisterUser(req.Login, hashPassword)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			logger.GetLoggerFromCtx(ctx).Warn(ctx,
				"user already exists",
				zap.String("login", req.Login),
				zap.Error(err))
			return nil, errs.ErrUserAlreadyExists
		}
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to register user",
			zap.String("login", req.Login),
			zap.Error(err))
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"user registered successfully",
		zap.String("login", req.Login),
		zap.String("user_id", userID))
	return &auth_service.RegisterResponse{UserId: userID}, nil
}

func (s *AuthService) Login(ctx context.Context, req *auth_service.LoginRequest) (*auth_service.LoginResponse, error) {
	if req.GetLogin() == "" {
		return nil, errs.ErrEmptyLogin
	}
	if req.GetPassword() == "" {
		return nil, errs.ErrEmptyPassword
	}

	userID, hash, err := s.storage.LoginUser(req.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			logger.GetLoggerFromCtx(ctx).Warn(ctx,
				"user not found",
				zap.String("login", req.Login),
				zap.Error(err))
			return nil, err
		}
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to login user",
			zap.String("login", req.Login),
			zap.Error(err))
		return nil, fmt.Errorf("failed to login user: %w", err)
	}

	ok := utils.CompareHash(req.Password, hash)
	if !ok {
		logger.GetLoggerFromCtx(ctx).Warn(ctx,
			"passwords do not match",
			zap.String("login", req.Login),
			zap.Error(err))
		return nil, errs.ErrWrongPassword
	}

	accessToken, err := utils.GenerateJWT(userID, s.jwtSecret, false, s.AccessExpiration)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to generate access jwt token",
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate access jwt token: %w", err)
	}
	refreshToken, err := utils.GenerateJWT(userID, s.jwtSecret, true, s.RefreshExpiration)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to generate refresh jwt token",
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh jwt token: %w", err)
	}
	err = s.cache.SaveToken(accessToken, userID, false)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to save access token in cache",
			zap.Error(err))
		return nil, fmt.Errorf("failed to save access token in cache: %w", err)
	}

	err = s.cache.SaveToken(refreshToken, userID, true)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to save refresh token in cache",
			zap.Error(err))
		return nil, fmt.Errorf("failed to save refresh token in cache: %w", err)
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"user logged in successfully",
		zap.String("login", req.Login),
		zap.String("user_id", userID),
		zap.String("access_token", accessToken))
	return &auth_service.LoginResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *auth_service.RefreshRequest) (*auth_service.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, errs.ErrInvalidToken
	}

	sub, isRefresh, expTime, err := utils.ParseJWT(req.RefreshToken, s.jwtSecret)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse refresh jwt token",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err))
		return nil, err
	}

	if !isRefresh {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"refresh token is not a refresh token",
			zap.String("refresh_token", req.RefreshToken),
			zap.String("sub", sub))
		return nil, errs.ErrInvalidToken
	}

	if time.Now().After(expTime) {
		logger.GetLoggerFromCtx(ctx).Warn(ctx,
			"refresh_token expired",
			zap.String("refresh_token", req.RefreshToken),
			zap.Time("exp_time", expTime))
		return nil, errs.ErrTokenExpired
	}

	userID, err := s.cache.GetToken(req.RefreshToken, true)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to get refresh_token from cache",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get refresh_token from cache: %w", err)
	}

	if userID != sub {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"refresh token user id does not match",
			zap.String("refresh_token", req.RefreshToken),
			zap.String("user_id", userID),
			zap.String("sub", sub))
		return nil, errs.ErrInvalidToken
	}

	accessToken, err := utils.GenerateJWT(userID, s.jwtSecret, false, s.AccessExpiration)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to generate access jwt token",
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate access jwt token: %w", err)
	}

	refreshToken, err := utils.GenerateJWT(userID, s.jwtSecret, true, s.RefreshExpiration)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to generate refresh jwt token",
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh jwt token: %w", err)
	}

	err = s.cache.SaveToken(accessToken, userID, false)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to save access token in cache",
			zap.Error(err))
		return nil, fmt.Errorf("failed to save access token in cache: %w", err)
	}

	err = s.cache.SaveToken(refreshToken, userID, true)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to save refresh token in cache",
			zap.Error(err))
		return nil, fmt.Errorf("failed to save refresh token in cache: %w", err)
	}

	err = s.cache.DeleteToken(req.RefreshToken, true)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to delete old refresh token from cache",
			zap.String("refresh_token", req.RefreshToken),
			zap.Error(err))
		return nil, fmt.Errorf("failed to delete old refresh token from cache: %w", err)
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"user refreshed tokens successfully",
		zap.String("user_id", userID),
		zap.String("access_token", accessToken))
	return &auth_service.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

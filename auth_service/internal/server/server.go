package server

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/auth_service/internal/ports"
	"github.com/jaam8/web_calculator/auth_service/internal/service"
	"github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/common-lib/grpc/interceptors"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"
)

func CreateGRPC(grpcSrv *service.AuthService) (*grpc.Server, error) {
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptors.AddLogMiddleware))
	auth_service.RegisterAuthServiceServer(server, grpcSrv)
	return server, nil
}

func NewAuthService(storage ports.StorageAdapter, cache ports.CacheAdapter,
	jwtSecret string, refreshExpiration, accessExpiration time.Duration) *service.AuthService {
	return service.NewAuthService(
		storage, cache, jwtSecret,
		refreshExpiration, accessExpiration)
}

func RunGRPC(ctx context.Context, server *grpc.Server, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx,
			"AUTH_SERVICE failed to create listener on port",
			zap.Int("port", port),
			zap.Error(err))
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("AUTH_SERVICE listening at :%d", port))
	if err = server.Serve(lis); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx,
			"AUTH_SERVICE failed to serve grpc server",
			zap.Error(err))
	}
}

package interceptors

import (
	"context"
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func AddLogMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx, _ = logger.New(ctx)
	ctx = context.WithValue(ctx, logger.KeyForRequestID, uuid.New().String())
	logger.GetLoggerFromCtx(ctx).Info(ctx, "gRPC request",
		zap.String("method", info.FullMethod),
		zap.Time("request time", time.Now()),
	)
	reply, err := handler(ctx, req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Warn(ctx, "gRPC hanler returned an error", zap.Error(err))
	}
	return reply, err
}

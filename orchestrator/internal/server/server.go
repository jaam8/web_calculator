package server

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/grpc/interceptors"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/ports"
	"github.com/jaam8/web_calculator/orchestrator/internal/service"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

func CreateGRPC(grpcSrv *service.OrchestratorService) (*grpc.Server, error) {
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptors.AddLogMiddleware))
	orchestrator.RegisterOrchestratorServiceServer(server, grpcSrv)
	return server, nil
}

func NewOrchestratorService(
	storage ports.StorageAdapter,
	expressionManager types.ExpressionManager) *service.OrchestratorService {
	return service.NewOrchestratorService(storage, expressionManager)
}

func RunGRPC(ctx context.Context, server *grpc.Server, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx,
			"ORCHESTRATOR failed to create listener on port",
			zap.Int("port", port),
			zap.Error(err))
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("ORCHESTRATOR listening at :%d", port))
	if err = server.Serve(lis); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx,
			"ORCHESTRATOR failed to serve grpc server",
			zap.Error(err))
	}
}

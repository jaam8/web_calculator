package main

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/grpc/pool"
	"github.com/jaam8/web_calculator/common-lib/logger"
	_ "github.com/jaam8/web_calculator/gateway/docs"
	"github.com/jaam8/web_calculator/gateway/internal/config"
	"github.com/jaam8/web_calculator/gateway/internal/http/handlers"
	"github.com/jaam8/web_calculator/gateway/internal/http/middlewares"
	"github.com/jaam8/web_calculator/gateway/internal/ports/adapters/orchestrator_adapters"
	"github.com/jaam8/web_calculator/gateway/internal/services"
	"github.com/labstack/echo/v4"
	swagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"time"
)

// @title Web Calculator API
// @version 1.0
// @description Web Calculator Gateway Service API
// @host localhost:8080
// @BasePath /api/v1
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx = context.WithValue(ctx, "log_level", cfg.LogLevel)
	ctx, _ = logger.New(ctx)

	orchestratorCfg := cfg.Orchestrator
	gatewayCfg := cfg.Gateway
	grpcPoolCfg := cfg.GrpcPool

	//region orchestrator grpc pool
	var orchestratorGrpcPool *pool.GrpcPool

	orchestratorAddress := fmt.Sprintf("%s:%d", orchestratorCfg.UpstreamName, orchestratorCfg.UpstreamPort)

	orchestratorGrpcPool, err = pool.NewGrpcPool(ctx, pool.Config{
		Address:        orchestratorAddress,
		MaxConnections: grpcPoolCfg.MaxConns,
		MinConnections: grpcPoolCfg.MinConns,
		DialOptions:    []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	})
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "couldn't create grpc pool for user service",
			zap.Int("MaxConnections", grpcPoolCfg.MaxConns),
			zap.Int("MinConnections", grpcPoolCfg.MinConns),
			zap.String("Address", orchestratorAddress),
			zap.Error(err),
		)
	}
	//endregion

	orchestratorAdapter := orchestrator_adapters.NewOrchestratorAdapter(orchestratorGrpcPool)
	orchestratorService := services.NewOrchestratorService(
		orchestratorAdapter,
		grpcPoolCfg.MaxRetries,
		time.Millisecond*time.Duration(grpcPoolCfg.BaseRetryDelayMs),
	)
	orchestratorHandler := handlers.NewOrchestratorHandler(orchestratorService)

	e := echo.New()

	e.Use(middlewares.CORSMiddleware)
	e.Use(middlewares.LogMiddleware)

	apiV1 := e.Group("/api/v1")

	apiV1.POST("/calculate", orchestratorHandler.Calculate)
	apiV1.GET("/expressions", orchestratorHandler.Expressions)
	apiV1.GET("/expressions/:id", orchestratorHandler.ExpressionByID)

	e.GET("/swagger/*", swagger.WrapHandler)

	go func() {
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			fmt.Sprintf("GATEWAY server started on port %d", gatewayCfg.Port))
		err = e.Start(fmt.Sprintf(":%d", gatewayCfg.Port))
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "GATEWAY server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()
	err = e.Shutdown(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "GATEWAY server failed to shutdown", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "GATEWAY server stopped")
}

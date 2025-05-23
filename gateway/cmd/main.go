package main

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/grpc/pool"
	"github.com/jaam8/web_calculator/common-lib/logger"
	_ "github.com/jaam8/web_calculator/gateway/docs"
	"github.com/jaam8/web_calculator/gateway/internal/config"
	grpc_servises "github.com/jaam8/web_calculator/gateway/internal/delivery/grpc"
	"github.com/jaam8/web_calculator/gateway/internal/delivery/http/handlers"
	"github.com/jaam8/web_calculator/gateway/internal/delivery/http/middlewares"
	"github.com/jaam8/web_calculator/gateway/internal/ports/adapters/auth_service_adapters"

	//"github.com/jaam8/web_calculator/gateway/internal/delivery/grpc"
	"github.com/jaam8/web_calculator/gateway/internal/ports/adapters/orchestrator_adapters"
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

	authCfg := cfg.AuthService
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
	orchestratorService := grpc_servises.NewOrchestratorService(
		orchestratorAdapter,
		grpcPoolCfg.MaxRetries,
		time.Millisecond*time.Duration(grpcPoolCfg.BaseRetryDelayMs),
	)
	orchestratorHandler := handlers.NewOrchestratorHandler(orchestratorService)

	// region auth_service grpc pool
	var authServiceGrpcPool *pool.GrpcPool
	authServiceAddress := fmt.Sprintf("%s:%d", authCfg.UpstreamName, authCfg.UpstreamPort)
	authServiceGrpcPool, err = pool.NewGrpcPool(ctx, pool.Config{
		Address:        authServiceAddress,
		MaxConnections: grpcPoolCfg.MaxConns,
		MinConnections: grpcPoolCfg.MinConns,
		DialOptions:    []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	})
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "couldn't create grpc pool for auth service",
			zap.Int("MaxConnections", grpcPoolCfg.MaxConns),
			zap.Int("MinConnections", grpcPoolCfg.MinConns),
			zap.String("Address", authServiceAddress),
			zap.Error(err),
		)
	}
	// endregion

	authServiceAdapter := auth_service_adapters.NewAuthServiceAdapter(authServiceGrpcPool)
	authService := grpc_servises.NewAuthService(
		authServiceAdapter,
		grpcPoolCfg.MaxRetries,
		time.Millisecond*time.Duration(grpcPoolCfg.BaseRetryDelayMs),
	)
	authHandler := handlers.NewAuthServiceHandler(
		authService,
		time.Duration(cfg.AccessTTL)*time.Minute,
		time.Duration(cfg.RefreshTTL)*time.Hour,
	)

	e := echo.New()

	apiV1 := e.Group("/api/v1")
	auth := apiV1.Group("/", middlewares.AuthMiddleware(cfg.JwtSecret))

	e.Use(middlewares.CORSMiddleware)
	e.Use(middlewares.LogMiddleware)

	auth.POST("calculate", orchestratorHandler.Calculate)
	auth.GET("expressions", orchestratorHandler.Expressions)
	auth.GET("expressions/:id", orchestratorHandler.ExpressionByID)
	apiV1.POST("/refresh-token", authHandler.Refresh)
	apiV1.POST("/login", authHandler.Login)
	apiV1.POST("/register", authHandler.Register)

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

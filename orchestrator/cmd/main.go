package main

import (
	"context"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/config"
	"github.com/jaam8/web_calculator/orchestrator/internal/server"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/utils"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, _ = logger.New(ctx)

	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx,
			"failed to load config", zap.Error(err))
	}

	orchestratorCfg := cfg.Orchestrator

	durations := map[string]int{
		"+": orchestratorCfg.TimeAddition,
		"-": orchestratorCfg.TimeSubtraction,
		"*": orchestratorCfg.TimeMultiplications,
		"/": orchestratorCfg.TimeDivisions,
	}
	expressionManager := utils.NewExpressionManager(durations)

	Server := server.NewOrchestratorService(expressionManager)
	grpcServer, err := server.CreateGRPC(Server)
	if err != nil {
		log.Fatalf("failed to create gRPC server: %v", err)
	}

	go server.RunGRPC(ctx, grpcServer, orchestratorCfg.Port)

	<-ctx.Done()
	grpcServer.GracefulStop()
	stop()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "ORCHESTRATOR server stopped")
}

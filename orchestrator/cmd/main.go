package main

import (
	"context"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/config"
	"github.com/jaam8/web_calculator/orchestrator/internal/server"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/utils"
	"log"
	"os"
	"os/signal"
)

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

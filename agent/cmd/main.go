package main

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/agent/internal/config"
	"github.com/jaam8/web_calculator/agent/internal/ports/adapters/orchestrator_adapters"
	"github.com/jaam8/web_calculator/agent/internal/server"
	"github.com/jaam8/web_calculator/agent/internal/service"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/signal"
	"time"
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
	agentCfg := cfg.Agent

	orchestratorAdapter := orchestrator_adapters.NewOrchestratorAdapter(
		fmt.Sprintf("%s:%d", orchestratorCfg.Host, orchestratorCfg.Port),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		time.Millisecond*time.Duration(orchestratorCfg.Timeout),
		orchestratorCfg.MaxRetries,
		time.Second*time.Duration(orchestratorCfg.BaseRetryDelay),
	)

	agentService := service.NewAgentService(orchestratorAdapter)
	server.RunAgentService(agentService, agentCfg.ComputingPower, agentCfg.WaitTime)

	select {
	case <-ctx.Done():
		logger.GetLoggerFromCtx(ctx).Info(ctx, "AGENTS stopped")
	}
}

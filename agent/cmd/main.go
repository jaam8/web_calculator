package main

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/agent/internal/config"
	"github.com/jaam8/web_calculator/agent/internal/ports/adapters/orchestrator_adapters"
	"github.com/jaam8/web_calculator/agent/internal/server"
	"github.com/jaam8/web_calculator/agent/internal/service"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"time"
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
	agentCfg := cfg.Agent

	orchestratorAdapter := orchestrator_adapters.NewOrchestratorAdapter(
		fmt.Sprintf("%s:%d", orchestratorCfg.UpstreamName, orchestratorCfg.UpstreamPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		time.Millisecond*time.Duration(orchestratorCfg.Timeout),
		orchestratorCfg.MaxRetries,
		time.Second*time.Duration(orchestratorCfg.BaseRetryDelay),
	)

	agentService := service.NewAgentService(orchestratorAdapter)
	server.RunAgentService(ctx, agentService, agentCfg.ComputingPower, agentCfg.WaitTime)

	select {
	case <-ctx.Done():
		logger.GetLoggerFromCtx(ctx).Info(ctx, "AGENTS stopped")
	}
}

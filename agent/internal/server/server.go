package server

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/agent/internal/service"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"sync"
)

func RunAgentService(ctx context.Context, agentService *service.AgentService, computingPower int, waitTime int) {
	var wg sync.WaitGroup

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			fmt.Sprintf("AGENT Starting worker %d", i))
		go func() {
			defer wg.Done()
			agentService.Work(ctx, waitTime)
		}()
	}
	wg.Wait()
}

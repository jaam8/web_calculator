package server

import (
	"github.com/jaam8/web_calculator/agent/internal/service"
	"log"
	"sync"
)

func RunAgentService(agentService *service.AgentService, computingPower int, waitTime int) {
	var wg sync.WaitGroup

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		log.Println("AGENT Starting worker ", i)
		go func() {
			defer wg.Done()
			agentService.Work(waitTime)
		}()
	}
	wg.Wait()
}

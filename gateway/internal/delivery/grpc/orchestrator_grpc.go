package grpc

import (
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/callers"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/gateway/internal/ports"
	"time"
)

type OrchestratorService struct {
	orchestratorAdapter *ports.OrchestratorAdapter
	MaxRetries          uint
	BaseDelay           time.Duration
}

func NewOrchestratorService(orchestratorAdapter ports.OrchestratorAdapter,
	maxRetries uint, baseDelay time.Duration,
) *OrchestratorService {
	return &OrchestratorService{
		orchestratorAdapter: &orchestratorAdapter,
		MaxRetries:          maxRetries,
		BaseDelay:           baseDelay,
	}
}

func (s *OrchestratorService) Calculate(request *orchestrator.CalculateRequest) (*orchestrator.CalculateResponse, error) {

	resultChan := make(chan *orchestrator.CalculateResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.orchestratorAdapter).Calculate(request)
		if err != nil {
			return fmt.Errorf("error in retry Calculate caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call Calculate: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

func (s *OrchestratorService) Expressions(request *orchestrator.ExpressionsRequest) (*orchestrator.ExpressionsResponse, error) {
	resultChan := make(chan *orchestrator.ExpressionsResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.orchestratorAdapter).Expressions(request)
		if err != nil {
			return fmt.Errorf("error in retry Expressions caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call Expressions: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

func (s *OrchestratorService) ExpressionByID(request *orchestrator.ExpressionByIdRequest) (*orchestrator.ExpressionByIdResponse, error) {
	resultChan := make(chan *orchestrator.ExpressionByIdResponse, 1)

	err := callers.Retry(func() error {
		response, err := (*s.orchestratorAdapter).ExpressionByID(request)
		if err != nil {
			return fmt.Errorf("error in retry ExpressionByID caller: %w", err)
		}
		resultChan <- response
		return nil
	}, s.MaxRetries, s.BaseDelay)

	if err != nil {
		return nil, fmt.Errorf("couldn't call ExpressionByID: %w", err)
	}

	response := <-resultChan
	close(resultChan)

	return response, nil
}

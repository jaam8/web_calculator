package orchestrator_adapters

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/grpc/pool"
)

type OrchestratorAdapter struct {
	grpcPool *pool.GrpcPool
}

func NewOrchestratorAdapter(grpcPool *pool.GrpcPool) *OrchestratorAdapter {
	return &OrchestratorAdapter{
		grpcPool: grpcPool,
	}
}

func (o OrchestratorAdapter) Calculate(request *orchestrator.CalculateRequest) (*orchestrator.CalculateResponse, error) {
	conn, err := o.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer o.grpcPool.Restore(conn) //nolint
	client := orchestrator.NewOrchestratorServiceClient(conn)
	response, grpcErr := client.Calculate(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in Calculate grpc: %w", grpcErr)
	}
	return response, nil
}

func (o OrchestratorAdapter) Expressions(request *orchestrator.ExpressionsRequest) (*orchestrator.ExpressionsResponse, error) {
	conn, err := o.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer o.grpcPool.Restore(conn) //nolint
	client := orchestrator.NewOrchestratorServiceClient(conn)
	response, grpcErr := client.Expressions(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in Expressions grpc: %w", grpcErr)
	}
	return response, nil
}

func (o OrchestratorAdapter) ExpressionByID(request *orchestrator.ExpressionByIdRequest) (*orchestrator.ExpressionByIdResponse, error) {
	conn, err := o.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer o.grpcPool.Restore(conn) //nolint
	client := orchestrator.NewOrchestratorServiceClient(conn)
	response, grpcErr := client.ExpressionById(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in ExpressionById grpc: %w", grpcErr)
	}
	return response, nil
}

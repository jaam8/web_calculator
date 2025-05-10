package orchestrator_adapters

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/agent/internal/models"
	"github.com/jaam8/web_calculator/common-lib/callers"
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"time"
)

type OrchestratorAdapter struct {
	Address        string
	DialOptions    []grpc.DialOption
	Timeout        time.Duration
	MaxRetries     uint
	BaseRetryDelay time.Duration
}

func NewOrchestratorAdapter(
	address string,
	dialOptions []grpc.DialOption,
	timeout time.Duration,
	maxRetries uint,
	baseRetryDelay time.Duration) *OrchestratorAdapter {
	return &OrchestratorAdapter{
		Address:        address,
		DialOptions:    dialOptions,
		Timeout:        timeout,
		MaxRetries:     maxRetries,
		BaseRetryDelay: baseRetryDelay,
	}
}

func (o *OrchestratorAdapter) GetGRPCClient() (*grpc.ClientConn, *orchestrator.OrchestratorServiceClient, error) {
	grpcConn, err := grpc.NewClient(o.Address, o.DialOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect to orchestrator by gRPC: %v", err)
	}

	client := orchestrator.NewOrchestratorServiceClient(grpcConn)

	return grpcConn, &client, nil
}

func (s *OrchestratorAdapter) GetTask() (models.Task, error) {
	conn, clientPointer, err := s.GetGRPCClient()
	if err != nil {
		return models.Task{}, fmt.Errorf("cannot connect to orchestrator by gRPC: %v", err)
	}

	defer conn.Close() //nolint

	resultChan := make(chan *orchestrator.Task, 1)

	err = callers.Retry(func() error {
		response, grpcErr := (*clientPointer).GetTask(context.Background(), &emptypb.Empty{})
		if grpcErr != nil {
			return fmt.Errorf("error in timeout gRPC caller: %v", grpcErr)
		}
		resultChan <- response.GetTask()
		return nil
	}, s.MaxRetries, s.BaseRetryDelay)

	if err != nil {
		log.Printf("error in timeout gRPC caller: %v", err)
		return models.Task{}, fmt.Errorf("couldn't get orchestrator.GetTask gRPC response: %v", err)
	}

	responseTask := <-resultChan
	close(resultChan)
	task := models.Task{
		ExpressionID:  int(responseTask.ExpressionId),
		TaskID:        int(responseTask.Id),
		Arg1:          responseTask.Arg1,
		Arg2:          responseTask.Arg2,
		Operation:     responseTask.Operation,
		OperationTime: responseTask.OperationTime.AsDuration(),
	}
	return task, nil
}

func (s *OrchestratorAdapter) ResultTask(
	expressionID, taskID int, result float64,
) (string, error) {
	conn, clientPointer, err := s.GetGRPCClient()
	if err != nil {
		return "", fmt.Errorf("cannot connect to orchestrator by gRPC: %v", err)
	}
	defer conn.Close() //nolint

	resultChan := make(chan string, 1)

	err = callers.Retry(func() error {
		request := &orchestrator.ResultTaskRequest{
			ExpressionId: int64(expressionID),
			Id:           int64(taskID),
			Result:       result,
		}
		response, grpcErr := (*clientPointer).ResultTask(context.Background(), request)
		if grpcErr != nil {
			return fmt.Errorf("error in timeout gRPC caller: %v", grpcErr)
		}
		resultChan <- response.GetStatus()
		return nil
	}, s.MaxRetries, s.BaseRetryDelay)

	if err != nil {
		log.Printf("error in timeout gRPC caller: %v", err)
		return "", fmt.Errorf("couldn't get orchestrator.ResultTaskRequest gRPC response: %v", err)
	}

	status := <-resultChan
	close(resultChan)
	if status == "task not found" {
		return "", errors.ErrTaskNotFound
	}
	if status == "task completed" {
		return status, nil
	}
	return "", errors.ErrInternalServerError
}

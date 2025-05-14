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

func (o *OrchestratorAdapter) GetTask() (models.Task, error) {
	conn, clientPointer, err := o.GetGRPCClient()
	if err != nil {
		return models.Task{}, fmt.Errorf("cannot connect to orchestrator by gRPC: %v", err)
	}

	defer conn.Close() //nolint

	resultChan := make(chan *orchestrator.Task, 1)

	err = callers.Timeout(func() error {
		response, grpcErr := (*clientPointer).GetTask(context.Background(), &emptypb.Empty{})
		if grpcErr != nil {
			return fmt.Errorf("error in timeout gRPC caller: %v", grpcErr)
		}
		resultChan <- response.GetTask()
		return nil
	}, o.Timeout)

	if err != nil {
		log.Printf("error in timeout gRPC caller: %v", err)
		return models.Task{}, fmt.Errorf("couldn't get orchestrator.GetTask gRPC response: %v", err)
	}

	responseTask := <-resultChan
	close(resultChan)
	task := models.Task{
		ExpressionID:  responseTask.GetExpressionId(),
		TaskID:        int(responseTask.GetId()),
		Arg1:          responseTask.GetArg1(),
		Arg2:          responseTask.GetArg2(),
		Operation:     responseTask.GetOperation(),
		OperationTime: responseTask.GetOperationTime().AsDuration(),
	}
	return task, nil
}

func (o *OrchestratorAdapter) ResultTask(
	expressionID string, taskID int, result float64,
) (string, error) {
	conn, clientPointer, err := o.GetGRPCClient()
	if err != nil {
		return "", fmt.Errorf("cannot connect to orchestrator by gRPC: %v", err)
	}
	defer conn.Close() //nolint

	resultChan := make(chan string, 1)
	err = callers.Retry(func() error {
		err = callers.Timeout(func() error {
			request := &orchestrator.ResultTaskRequest{
				ExpressionId: expressionID,
				Id:           int64(taskID),
				Result:       result,
			}
			response, grpcErr := (*clientPointer).ResultTask(context.Background(), request)
			if grpcErr != nil {
				return fmt.Errorf("error in timeout gRPC caller: %v", grpcErr)
			}
			resultChan <- response.GetStatus()
			return nil
		}, o.Timeout)
		if err != nil {
			return fmt.Errorf("error in retry gRPC caller: %v", err)
		}
		return nil
	}, o.MaxRetries, o.BaseRetryDelay)

	if err != nil {
		log.Printf("error in timeout gRPC caller: %v", err)
		return "", fmt.Errorf("couldn't get orchestrator.ResultTaskRequest gRPC response: %v", err)
	}

	status := <-resultChan
	close(resultChan)
	if status == "task not found" {
		return "", errors.ErrTaskNotFound
	}

	return status, nil
}

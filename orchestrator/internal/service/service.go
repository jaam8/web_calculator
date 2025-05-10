package service

import (
	"context"
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/helper"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/utils"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrchestratorService struct {
	orchestrator.OrchestratorServiceServer
	expressionManager types.ExpressionManager
}

func NewOrchestratorService(expressionManager types.ExpressionManager) *OrchestratorService {
	return &OrchestratorService{
		expressionManager: expressionManager,
	}
}

func (s *OrchestratorService) Calculate(
	ctx context.Context, request *orchestrator.CalculateRequest,
) (*orchestrator.CalculateResponse, error) {
	rpn, err := helper.ToRPN(request.Expression)
	if err != nil {
		return nil, err
	}
	expressionID, err := s.expressionManager.CreateExpression()
	if err != nil {
		return nil, err
	}
	taskManager, exists := s.expressionManager.GetTaskManager(expressionID)
	if !exists {
		return nil, errors.ErrInternalServerError
	}
	go utils.Process(rpn, taskManager, s.expressionManager, expressionID)
	return &orchestrator.CalculateResponse{Id: int64(expressionID)}, nil
}

func (s *OrchestratorService) Expressions(
	ctx context.Context, request *orchestrator.ExpressionsRequest,
) (*orchestrator.ExpressionsResponse, error) {
	expressions := s.expressionManager.GetExpressions()
	exprs := make([]*orchestrator.Expression, len(expressions))
	for i, e := range expressions {
		expr := &orchestrator.Expression{
			Id:     int64(e.ExpressionID),
			Status: e.Status,
			Result: e.Result,
		}
		exprs[i] = expr
	}
	return &orchestrator.ExpressionsResponse{Expressions: exprs}, nil
}

func (s *OrchestratorService) ExpressionByID(
	ctx context.Context, request *orchestrator.ExpressionByIdRequest,
) (*orchestrator.ExpressionByIdResponse, error) {
	expression, ok := s.expressionManager.GetExpression(int(request.Id))
	if !ok {
		// add norm error
		return nil, errors.ErrExpressionNotFound
	}
	expr := &orchestrator.Expression{
		Id:     int64(expression.ExpressionID),
		Status: expression.Status,
		Result: expression.Result,
	}
	return &orchestrator.ExpressionByIdResponse{Expression: expr}, nil
}

func (s *OrchestratorService) ResultTask(
	ctx context.Context, request *orchestrator.ResultTaskRequest,
) (*orchestrator.ResultTaskResponse, error) {
	// todo why task id and not expression id
	taskManager, ok := s.expressionManager.GetTaskManager(int(request.ExpressionId))
	if !ok {
		return nil, errors.ErrTaskNotFound

	}
	result := models.Result{
		ExpressionID: int(request.ExpressionId),
		TaskID:       int(request.Id),
		Result:       request.Result,
	}
	taskManager.AddResult(result)
	return &orchestrator.ResultTaskResponse{Status: "task completed"}, nil
}

func (s *OrchestratorService) GetTask(
	ctx context.Context, request *emptypb.Empty,
) (*orchestrator.GetTaskResponse, error) {
	select {
	case task := <-s.expressionManager.GetTasks():
		if task.TaskID != 0 {
			return &orchestrator.GetTaskResponse{
				Task: &orchestrator.Task{
					Id:            int64(task.TaskID),
					Arg1:          task.Arg1,
					Arg2:          task.Arg2,
					Operation:     task.Operation,
					OperationTime: durationpb.New(task.OperationTime),
				},
			}, nil
		}
	default:
		return nil, errors.ErrTaskNotFound
	}
	return nil, errors.ErrInternalServerError
}

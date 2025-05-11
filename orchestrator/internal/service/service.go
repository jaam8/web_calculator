package service

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/helper"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/utils"
	"go.uber.org/zap"
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
	logger.GetLoggerFromCtx(ctx).Debug(ctx,
		fmt.Sprintf("RPN for expression: %s", request.Expression),
		zap.Any("rpn", rpn),
		zap.Error(err))
	if err != nil {
		return nil, err
	}
	expressionID, err := s.expressionManager.CreateExpression()
	logger.GetLoggerFromCtx(ctx).Debug(ctx,
		fmt.Sprintf("Created expression with id: %d", expressionID),
		zap.Int("expressionID", expressionID),
		zap.Error(err),
	)
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
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		fmt.Sprintf("got %d expressions", len(expressions)),
		zap.Int("expressionsCount", len(expressions)))
	return &orchestrator.ExpressionsResponse{Expressions: exprs}, nil
}

func (s *OrchestratorService) ExpressionById(
	ctx context.Context, request *orchestrator.ExpressionByIdRequest,
) (*orchestrator.ExpressionByIdResponse, error) {
	expression, ok := s.expressionManager.GetExpression(int(request.Id))
	if !ok {
		logger.GetLoggerFromCtx(ctx).Warn(ctx,
			fmt.Sprintf("no expression found for id: %d", request.Id),
			zap.Int("expressionID", int(request.Id)),
		)
		return nil, errors.ErrExpressionNotFound
	}
	expr := &orchestrator.Expression{
		Id:     int64(expression.ExpressionID),
		Status: expression.Status,
		Result: expression.Result,
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"got expression by id",
		zap.Int64("ExpressionID", expr.Id),
		zap.String("status", expr.Status),
		zap.Float64p("result", expr.Result))
	return &orchestrator.ExpressionByIdResponse{Expression: expr}, nil
}

func (s *OrchestratorService) ResultTask(
	ctx context.Context, request *orchestrator.ResultTaskRequest,
) (*orchestrator.ResultTaskResponse, error) {
	taskManager, ok := s.expressionManager.GetTaskManager(int(request.ExpressionId))
	if !ok {
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("no task manager found for expression with id: %d", request.ExpressionId),
			zap.Int("expressionID", int(request.ExpressionId)),
		)
		return nil, errors.ErrTaskNotFound
	}
	result := models.Result{
		ExpressionID: int(request.ExpressionId),
		TaskID:       int(request.Id),
		Result:       request.Result,
	}
	taskManager.AddResult(result)
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		fmt.Sprintf("got result for task with id: %d", request.Id),
		zap.Int("expressionID", int(request.ExpressionId)),
		zap.Int("taskID", int(request.Id)),
		zap.Float64("result", request.Result),
	)
	return &orchestrator.ResultTaskResponse{Status: "task completed"}, nil
}

func (s *OrchestratorService) GetTask(
	ctx context.Context, _ *emptypb.Empty,
) (*orchestrator.GetTaskResponse, error) {
	select {
	case task := <-s.expressionManager.GetTasks():
		if task.TaskID != 0 {
			logger.GetLoggerFromCtx(ctx).Debug(ctx,
				fmt.Sprintf("send task with id: %d", task.TaskID),
				zap.Int("expressionID", task.ExpressionID),
				zap.Int("taskID", task.TaskID),
				zap.Float64("arg1", task.Arg1),
				zap.Float64("arg2", task.Arg2),
				zap.String("operation", task.Operation),
				zap.Duration("operationTime", task.OperationTime),
			)
			return &orchestrator.GetTaskResponse{
				Task: &orchestrator.Task{
					ExpressionId:  int64(task.ExpressionID),
					Id:            int64(task.TaskID),
					Arg1:          task.Arg1,
					Arg2:          task.Arg2,
					Operation:     task.Operation,
					OperationTime: durationpb.New(task.OperationTime),
				},
			}, nil
		}
	default:
		logger.GetLoggerFromCtx(ctx).Warn(ctx, "no task found")
		return nil, errors.ErrTaskNotFound
	}
	logger.GetLoggerFromCtx(ctx).Error(ctx,
		"internal server error",
		zap.Error(errors.ErrInternalServerError),
	)
	return nil, errors.ErrInternalServerError
}

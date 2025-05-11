package service

import (
	"context"
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
	"time"
)

type MockExpressionManager struct {
	mock.Mock
	tasks chan models.Task
}

func (m *MockExpressionManager) CreateExpression() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockExpressionManager) GetTaskManager(exprID int) (types.TaskManager, bool) {
	args := m.Called(exprID)
	return args.Get(0).(types.TaskManager), args.Bool(1)
}

func (m *MockExpressionManager) GetTasks() chan models.Task {
	return m.tasks
}

func (m *MockExpressionManager) GetExpressions() []*models.Expression {
	args := m.Called()
	return args.Get(0).([]*models.Expression)
}

func (m *MockExpressionManager) GetExpression(id int) (*models.Expression, bool) {
	args := m.Called(id)
	return args.Get(0).(*models.Expression), args.Bool(1)
}

func (m *MockExpressionManager) AddTask(task models.Task) {
	m.tasks <- task
}

func (m *MockExpressionManager) ExpressionDone(exprID int, res float64) {
}

func (m *MockExpressionManager) ExpressionError(exprID int) {
	m.Called(exprID)
}

type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) CreateTask(arg1, arg2 float64, oper string, exprID int) models.Task {
	return models.Task{}
}

func (m *MockTaskManager) GetResult() models.Result {
	return models.Result{}
}

func (m *MockTaskManager) AddResult(result models.Result) {
	m.Called(result)
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		mockSetup      func(exprManager *MockExpressionManager, taskManager *MockTaskManager)
		expectedErr    bool
		expectedRespID int64
	}{
		{
			name:       "success",
			expression: "3+4",
			mockSetup: func(exprManager *MockExpressionManager, taskManager *MockTaskManager) {
				exprManager.On("CreateExpression").Return(1, nil)
				exprManager.On("GetTaskManager", 1).Return(taskManager, true)
			},
			expectedErr:    false,
			expectedRespID: 1,
		},
		{
			name:        "invalid RPN expression",
			expression:  "",
			mockSetup:   func(_ *MockExpressionManager, _ *MockTaskManager) {},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprManager := new(MockExpressionManager)
			taskManager := new(MockTaskManager)
			exprManager.tasks = make(chan models.Task, 1)

			tt.mockSetup(exprManager, taskManager)
			service := NewOrchestratorService(exprManager)

			resp, err := service.Calculate(
				context.Background(),
				&orchestrator.CalculateRequest{
					Expression: tt.expression,
				},
			)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRespID, resp.Id)
			}
			exprManager.AssertExpectations(t)
		})
	}
}

func TestExpressionByID(t *testing.T) {
	tests := []struct {
		name        string
		inputID     int
		mockSetup   func(*MockExpressionManager)
		expectedErr error
	}{
		{
			name:    "not found",
			inputID: 999,
			mockSetup: func(exprManager *MockExpressionManager) {
				exprManager.On("GetExpression", 999).Return(&models.Expression{}, false)
			},
			expectedErr: errors.ErrExpressionNotFound,
		},
		{
			name:    "expression found",
			inputID: 1,
			mockSetup: func(exprManager *MockExpressionManager) {
				expr := &models.Expression{
					ExpressionID: 1,
					Result:       func(d float64) *float64 { return &d }(7),
				}
				exprManager.On("GetExpression", 1).Return(expr, true)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprManager := new(MockExpressionManager)
			tt.mockSetup(exprManager)
			service := NewOrchestratorService(exprManager)

			_, err := service.ExpressionById(
				context.Background(),
				&orchestrator.ExpressionByIdRequest{
					Id: int64(tt.inputID),
				},
			)
			assert.ErrorIs(t, err, tt.expectedErr)
			exprManager.AssertExpectations(t)
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name        string
		setupChan   func() chan models.Task
		expectedErr error
	}{
		{
			name: "empty task queue",
			setupChan: func() chan models.Task {
				return make(chan models.Task, 0)
			},
			expectedErr: errors.ErrTaskNotFound,
		},
		{
			name: "task available",
			setupChan: func() chan models.Task {
				ch := make(chan models.Task, 1)
				ch <- models.Task{
					TaskID:        42,
					Arg1:          3,
					Arg2:          4,
					Operation:     "+",
					OperationTime: time.Second,
				}
				return ch
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprManager := new(MockExpressionManager)
			exprManager.tasks = tt.setupChan()

			service := NewOrchestratorService(exprManager)

			_, err := service.GetTask(context.Background(), &emptypb.Empty{})
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestResultTask(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(exprMgr *MockExpressionManager, taskMgr *MockTaskManager)
		request        *orchestrator.ResultTaskRequest
		expectedError  error
		expectedStatus string
	}{
		{
			name: "task manager not found",
			setupMocks: func(exprMgr *MockExpressionManager, taskMgr *MockTaskManager) {
				exprMgr.On("GetTaskManager", 999).Return((*MockTaskManager)(nil), false)
			},
			request: &orchestrator.ResultTaskRequest{
				ExpressionId: 999,
				Id:           1,
				Result:       7.0,
			},
			expectedError: errors.ErrTaskNotFound,
		},
		{
			name: "success",
			setupMocks: func(exprMgr *MockExpressionManager, taskMgr *MockTaskManager) {
				exprMgr.On("GetTaskManager", 1).Return(taskMgr, true)
				taskMgr.On("AddResult", models.Result{
					ExpressionID: 1,
					TaskID:       1,
					Result:       7.0,
				}).Return()
			},
			request: &orchestrator.ResultTaskRequest{
				ExpressionId: 1,
				Id:           1,
				Result:       7.0,
			},
			expectedError:  nil,
			expectedStatus: "task completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprMgr := new(MockExpressionManager)
			taskMgr := new(MockTaskManager)

			tt.setupMocks(exprMgr, taskMgr)

			service := NewOrchestratorService(exprMgr)
			resp, err := service.ResultTask(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, resp.Status)
			}

			exprMgr.AssertExpectations(t)
			taskMgr.AssertExpectations(t)
		})
	}
}

func TestExpressions(t *testing.T) {
	exprMgr := new(MockExpressionManager)

	exprMgr.On("GetExpressions").Return([]*models.Expression{
		{
			ExpressionID: 1,
			Status:       "done",
			Result:       func(d float64) *float64 { return &d }(7),
		},
		{
			ExpressionID: 2,
			Status:       "pending",
		},
	})

	service := NewOrchestratorService(exprMgr)

	resp, err := service.Expressions(context.Background(), &orchestrator.ExpressionsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Expressions, 2)
	assert.Equal(t, int64(1), resp.Expressions[0].Id)
	assert.Equal(t, "done", resp.Expressions[0].Status)
	assert.Equal(t, func(d float64) *float64 { return &d }(7), resp.Expressions[0].Result)

	assert.Equal(t, int64(2), resp.Expressions[1].Id)
	assert.Equal(t, "pending", resp.Expressions[1].Status)

	exprMgr.AssertExpectations(t)
}

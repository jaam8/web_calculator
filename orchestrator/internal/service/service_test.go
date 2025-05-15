package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	"testing"
	"time"
)

type MockExpressionManager struct {
	mock.Mock
	tasks chan models.Task
	mutex sync.Mutex
}

func (m *MockExpressionManager) CreateExpression(expression *models.Expression) error {
	args := m.Called(expression)
	return args.Error(0)
}

func (m *MockExpressionManager) GetTaskManager(exprID uuid.UUID) (types.TaskManager, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	args := m.Called(exprID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.TaskManager), args.Error(1)
}

func (m *MockExpressionManager) GetTasks() chan models.Task {
	return m.tasks
}

func (m *MockExpressionManager) GetExpressions() []*models.Expression {
	args := m.Called()
	return args.Get(0).([]*models.Expression)
}

func (m *MockExpressionManager) GetExpression(id uuid.UUID) (*models.Expression, bool) {
	args := m.Called(id)
	return args.Get(0).(*models.Expression), args.Bool(1)
}

func (m *MockExpressionManager) AddTask(task models.Task) {
	m.tasks <- task
}

func (m *MockExpressionManager) ExpressionDone(exprID uuid.UUID, res float64) {
	m.Called(exprID, res)
}

func (m *MockExpressionManager) ExpressionError(exprID uuid.UUID) {
	m.Called(exprID)
}

type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) CreateTask(arg1, arg2 float64, oper string, exprID uuid.UUID) models.Task {
	args := m.Called(arg1, arg2, oper, exprID)
	return args.Get(0).(models.Task)
}

func (m *MockTaskManager) GetResult() models.Result {
	args := m.Called()
	return args.Get(0).(models.Result)
}

func (m *MockTaskManager) AddResult(result models.Result) {
	m.Called(result)
}

type MockStorageAdapter struct {
	mock.Mock
}

func (m *MockStorageAdapter) SaveExpression(expr models.Expression) (uuid.UUID, error) {
	args := m.Called(expr)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockStorageAdapter) GetExpressions(userID uuid.UUID) ([]*models.Expression, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Expression), args.Error(1)
}

func (m *MockStorageAdapter) GetExpressionById(userId uuid.UUID, id uuid.UUID) (*models.Expression, error) {
	args := m.Called(userId, id)
	return args.Get(0).(*models.Expression), args.Error(1)
}

func (m *MockStorageAdapter) UpdateExpression(userID, expressionID uuid.UUID, status *string, result *float64) error {
	args := m.Called(userID, expressionID, status, result)
	return args.Error(0)
}

// setupCommonMocks handles setting up mocks that may be needed across multiple tests due to goroutines
func setupCommonMocks(taskManager *MockTaskManager, exprManager *MockExpressionManager) {
	exprID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	taskManager.On("CreateTask", 3.0, 4.0, "+", exprID).Maybe().Return(models.Task{
		ExpressionID:  exprID,
		TaskID:        42,
		Arg1:          3,
		Arg2:          4,
		Operation:     "+",
		OperationTime: time.Second,
	})

	taskManager.On("GetResult").Maybe().Return(models.Result{
		ExpressionID: exprID,
		TaskID:       42,
		Result:       7.0,
	})

	// Mock for ExpressionDone
	exprManager.On("ExpressionDone", exprID, 7.0).Maybe().Return()
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		mockSetup      func(exprManager *MockExpressionManager, taskManager *MockTaskManager, storage *MockStorageAdapter)
		expectedErr    bool
		expectedRespID string
	}{
		{
			name:       "success",
			expression: "3+4",
			mockSetup: func(exprManager *MockExpressionManager, taskManager *MockTaskManager, storage *MockStorageAdapter) {
				exprID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
				status := "done"
				result := 7.0
				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)

				storage.On("SaveExpression", mock.AnythingOfType("models.Expression")).Return(exprID, nil)
				storage.On("UpdateExpression", userID, exprID, &status, &result).Return(nil)

				exprManager.On("CreateExpression", mock.AnythingOfType("*models.Expression")).Return(nil)
				exprManager.On("GetTaskManager", exprID).Return(taskManager, nil)
			},
			expectedErr:    false,
			expectedRespID: "00000000-0000-0000-0000-000000000001",
		},
		{
			name:       "invalid RPN expression",
			expression: "",
			mockSetup: func(_ *MockExpressionManager, taskManager *MockTaskManager, _ *MockStorageAdapter) {
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprManager := new(MockExpressionManager)
			taskManager := new(MockTaskManager)
			storage := new(MockStorageAdapter)
			exprManager.tasks = make(chan models.Task, 1)

			tt.mockSetup(exprManager, taskManager, storage)
			service := NewOrchestratorService(storage, exprManager)

			ctx := context.Background()
			ctx, _ = logger.New(ctx)

			resp, err := service.Calculate(
				ctx,
				&orchestrator.CalculateRequest{
					Expression: tt.expression,
					UserId:     "00000000-0000-0000-0000-000000000002", // valid UUID
				},
			)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRespID, resp.Id)
			}

			// Allow some time for goroutines to complete
			time.Sleep(100 * time.Millisecond)

			exprManager.AssertExpectations(t)
			storage.AssertExpectations(t)
			taskManager.AssertExpectations(t)
		})
	}
}

func TestExpressionByID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expressionID   string
		mockSetup      func(*MockStorageAdapter, *MockTaskManager, *MockExpressionManager)
		expectedErr    error
		expectedStatus string
		expectedResult *float64
	}{
		{
			name:         "not found",
			userID:       "00000000-0000-0000-0000-000000000001",
			expressionID: "00000000-0000-0000-0000-000000000099",
			mockSetup: func(storage *MockStorageAdapter, taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				exprID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
				storage.On("GetExpressionById", userID, exprID).Return(&models.Expression{}, errors.ErrExpressionNotFound)

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			expectedErr: errors.ErrExpressionNotFound,
		},
		{
			name:         "expression found",
			userID:       "00000000-0000-0000-0000-000000000001",
			expressionID: "00000000-0000-0000-0000-000000000002",
			mockSetup: func(storage *MockStorageAdapter, taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				exprID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
				result := 7.0
				expr := &models.Expression{
					ExpressionID: exprID,
					UserId:       userID,
					Status:       "done",
					Result:       &result,
				}
				storage.On("GetExpressionById", userID, exprID).Return(expr, nil)

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			expectedErr:    nil,
			expectedStatus: "done",
			expectedResult: func() *float64 { r := 7.0; return &r }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(MockStorageAdapter)
			exprManager := new(MockExpressionManager)
			taskManager := new(MockTaskManager)
			exprManager.tasks = make(chan models.Task, 1)

			tt.mockSetup(storage, taskManager, exprManager)
			service := NewOrchestratorService(storage, exprManager)

			ctx := context.Background()
			ctx, _ = logger.New(ctx)

			resp, err := service.ExpressionById(
				ctx,
				&orchestrator.ExpressionByIdRequest{
					UserId: tt.userID,
					Id:     tt.expressionID,
				},
			)

			assert.ErrorIs(t, err, tt.expectedErr)

			if tt.expectedErr == nil {
				assert.Equal(t, tt.expressionID, resp.Expression.Id)
				assert.Equal(t, tt.expectedStatus, resp.Expression.Status)
				assert.Equal(t, tt.expectedResult, resp.Expression.Result)
			}

			// Allow some time for goroutines to complete
			time.Sleep(100 * time.Millisecond)

			storage.AssertExpectations(t)
			taskManager.AssertExpectations(t)
			exprManager.AssertExpectations(t)
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(taskManager *MockTaskManager, exprManager *MockExpressionManager)
		setupChan   func() chan models.Task
		expectedErr error
	}{
		{
			name: "empty task queue",
			setupMocks: func(taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			setupChan: func() chan models.Task {
				return make(chan models.Task, 0)
			},
			expectedErr: errors.ErrTaskNotFound,
		},
		{
			name: "task available",
			setupMocks: func(taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			setupChan: func() chan models.Task {
				ch := make(chan models.Task, 1)
				ch <- models.Task{
					ExpressionID:  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
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
			taskManager := new(MockTaskManager)
			storage := new(MockStorageAdapter)

			exprManager.tasks = tt.setupChan()

			tt.setupMocks(taskManager, exprManager)

			service := NewOrchestratorService(storage, exprManager)

			ctx := context.Background()
			ctx, _ = logger.New(ctx)

			resp, err := service.GetTask(ctx, &emptypb.Empty{})
			task := resp.GetTask()
			assert.ErrorIs(t, err, tt.expectedErr)

			if tt.expectedErr == nil {
				assert.NotNil(t, task)
				assert.Equal(t, "00000000-0000-0000-0000-000000000001", task.ExpressionId)
				assert.Equal(t, int64(42), task.Id)
				assert.Equal(t, 3.0, task.Arg1)
				assert.Equal(t, 4.0, task.Arg2)
				assert.Equal(t, "+", task.Operation)
			}

			time.Sleep(100 * time.Millisecond)

			taskManager.AssertExpectations(t)
			exprManager.AssertExpectations(t)
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
				exprID := uuid.MustParse("00000000-0000-0000-0000-000000000999")
				exprMgr.On("GetTaskManager", exprID).Return(nil, errors.ErrTaskNotFound)

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskMgr, exprMgr)
			},
			request: &orchestrator.ResultTaskRequest{
				ExpressionId: "00000000-0000-0000-0000-000000000999",
				Id:           1,
				Result:       7.0,
			},
			expectedError: errors.ErrTaskNotFound,
		},
		{
			name: "success",
			setupMocks: func(exprMgr *MockExpressionManager, taskMgr *MockTaskManager) {
				exprID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				exprMgr.On("GetTaskManager", exprID).Return(taskMgr, nil)
				taskMgr.On("AddResult", mock.AnythingOfType("models.Result")).Return()

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskMgr, exprMgr)
			},
			request: &orchestrator.ResultTaskRequest{
				ExpressionId: "00000000-0000-0000-0000-000000000001",
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
			storage := new(MockStorageAdapter)
			exprMgr.tasks = make(chan models.Task, 1)

			tt.setupMocks(exprMgr, taskMgr)

			ctx := context.Background()
			ctx, _ = logger.New(ctx)

			service := NewOrchestratorService(storage, exprMgr)
			resp, err := service.ResultTask(ctx, tt.request)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, resp.Status)
			}

			// Allow some time for goroutines to complete
			time.Sleep(100 * time.Millisecond)

			exprMgr.AssertExpectations(t)
			taskMgr.AssertExpectations(t)
		})
	}
}

func TestExpressions(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name          string
		mockSetup     func(*MockStorageAdapter, *MockTaskManager, *MockExpressionManager)
		expectedCount int
		expectedError error
	}{
		{
			name: "get expressions success",
			mockSetup: func(storage *MockStorageAdapter, taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				expressions := []*models.Expression{
					{
						ExpressionID: uuid.MustParse("00000000-0000-0000-0000-000000000011"),
						UserId:       userID,
						Status:       "done",
						Result:       func() *float64 { r := 7.0; return &r }(),
					},
					{
						ExpressionID: uuid.MustParse("00000000-0000-0000-0000-000000000012"),
						UserId:       userID,
						Status:       "pending",
						Result:       func() *float64 { r := 3.0; return &r }(),
					},
				}
				storage.On("GetExpressions", userID).Return(expressions, nil)

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "no expressions found",
			mockSetup: func(storage *MockStorageAdapter, taskManager *MockTaskManager, exprManager *MockExpressionManager) {
				storage.On("GetExpressions", userID).Return([]*models.Expression{}, errors.ErrExpressionNotFound)

				// Set up common mocks that might be needed from goroutines
				setupCommonMocks(taskManager, exprManager)
			},
			expectedCount: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(MockStorageAdapter)
			exprMgr := new(MockExpressionManager)
			taskManager := new(MockTaskManager)
			exprMgr.tasks = make(chan models.Task, 1)

			tt.mockSetup(storage, taskManager, exprMgr)

			service := NewOrchestratorService(storage, exprMgr)

			ctx := context.Background()
			ctx, _ = logger.New(ctx)

			resp, err := service.Expressions(ctx, &orchestrator.ExpressionsRequest{
				UserId: userID.String(),
			})

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Expressions, tt.expectedCount)
			}

			// Allow some time for goroutines to complete
			time.Sleep(100 * time.Millisecond)

			storage.AssertExpectations(t)
			taskManager.AssertExpectations(t)
			exprMgr.AssertExpectations(t)
		})
	}
}

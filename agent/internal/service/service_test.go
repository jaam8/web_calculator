package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jaam8/web_calculator/agent/internal/models"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Мок для OrchestratorAdapter
type MockOrchestratorAdapter struct {
	mock.Mock
}

func (m *MockOrchestratorAdapter) GetTask() (models.Task, error) {
	args := m.Called()
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockOrchestratorAdapter) ResultTask(expressionID string, taskID int, result float64) (string, error) {
	args := m.Called(expressionID, taskID, result)
	return args.String(0), args.Error(1)
}

// Мок для логгера
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func TestDoTask(t *testing.T) {
	tests := []struct {
		name        string
		task        models.Task
		expected    float64
		expectedErr error
	}{
		{
			name: "Addition",
			task: models.Task{
				ExpressionID:  "expr1",
				TaskID:        1,
				Arg1:          10,
				Arg2:          5,
				Operation:     "+",
				OperationTime: 0, // Skip sleep in tests
			},
			expected:    15,
			expectedErr: nil,
		},
		{
			name: "Subtraction",
			task: models.Task{
				ExpressionID:  "expr2",
				TaskID:        2,
				Arg1:          10,
				Arg2:          5,
				Operation:     "-",
				OperationTime: 0,
			},
			expected:    5,
			expectedErr: nil,
		},
		{
			name: "Multiplication",
			task: models.Task{
				ExpressionID:  "expr3",
				TaskID:        3,
				Arg1:          10,
				Arg2:          5,
				Operation:     "*",
				OperationTime: 0,
			},
			expected:    50,
			expectedErr: nil,
		},
		{
			name: "Division",
			task: models.Task{
				ExpressionID:  "expr4",
				TaskID:        4,
				Arg1:          10,
				Arg2:          5,
				Operation:     "/",
				OperationTime: 0,
			},
			expected:    2,
			expectedErr: nil,
		},
		{
			name: "Division by zero",
			task: models.Task{
				ExpressionID:  "expr5",
				TaskID:        5,
				Arg1:          10,
				Arg2:          0,
				Operation:     "/",
				OperationTime: 0,
			},
			expected:    0,
			expectedErr: errs.ErrDivideByZero,
		},
		{
			name: "Invalid operation",
			task: models.Task{
				ExpressionID:  "expr6",
				TaskID:        6,
				Arg1:          10,
				Arg2:          5,
				Operation:     "%",
				OperationTime: 0,
			},
			expected:    0,
			expectedErr: errs.ErrInvalidExpression,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DoTask(tt.task)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgentService_GetTask(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockOrchestratorAdapter)
		expectedTask  models.Task
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("GetTask").Return(models.Task{
					ExpressionID:  "expr1",
					TaskID:        1,
					Arg1:          10,
					Arg2:          5,
					Operation:     "+",
					OperationTime: time.Millisecond * 100,
				}, nil)
			},
			expectedTask: models.Task{
				ExpressionID:  "expr1",
				TaskID:        1,
				Arg1:          10,
				Arg2:          5,
				Operation:     "+",
				OperationTime: time.Millisecond * 100,
			},
			expectedError: nil,
		},
		{
			name: "Task not found",
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("GetTask").Return(models.Task{}, errs.ErrTaskNotFound)
			},
			expectedTask:  models.Task{TaskID: 0},
			expectedError: errs.ErrTaskNotFound,
		},
		{
			name: "Other error",
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("GetTask").Return(models.Task{}, errors.New("connection error"))
			},
			expectedTask:  models.Task{},
			expectedError: errors.New("connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter := new(MockOrchestratorAdapter)
			tt.mockSetup(mockAdapter)

			service := NewAgentService(mockAdapter)
			task, err := service.GetTask()

			mockAdapter.AssertExpectations(t)

			// Check error behavior
			if tt.expectedError != nil {
				if errors.Is(tt.expectedError, errs.ErrTaskNotFound) {
					assert.ErrorIs(t, err, errs.ErrTaskNotFound)
				} else {
					assert.EqualError(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// Check task
			assert.Equal(t, tt.expectedTask, task)
		})
	}
}

func TestAgentService_ResultTask(t *testing.T) {
	tests := []struct {
		name          string
		result        models.Result
		mockSetup     func(*MockOrchestratorAdapter)
		expectedError error
	}{
		{
			name: "Success",
			result: models.Result{
				ExpressionID: "expr1",
				TaskID:       1,
				Result:       15,
			},
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("ResultTask", "expr1", 1, 15.0).Return("ok", nil)
			},
			expectedError: nil,
		},
		{
			name: "Task not found",
			result: models.Result{
				ExpressionID: "expr2",
				TaskID:       2,
				Result:       20,
			},
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("ResultTask", "expr2", 2, 20.0).Return("", errs.ErrTaskNotFound)
			},
			expectedError: errs.ErrTaskNotFound,
		},
		{
			name: "Other error",
			result: models.Result{
				ExpressionID: "expr3",
				TaskID:       3,
				Result:       30,
			},
			mockSetup: func(m *MockOrchestratorAdapter) {
				m.On("ResultTask", "expr3", 3, 30.0).Return("", errors.New("connection error"))
			},
			expectedError: errors.New("connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter := new(MockOrchestratorAdapter)
			tt.mockSetup(mockAdapter)

			service := NewAgentService(mockAdapter)
			err := service.ResultTask(tt.result)

			mockAdapter.AssertExpectations(t)
			fmt.Printf("%v", err)
			fmt.Printf("%v", errs.ErrInvalidToken)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAgentService_Work с помощью контекста с таймаутом
func TestAgentService_Work(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockAdapter := new(MockOrchestratorAdapter)

	mockAdapter.On("GetTask").Return(models.Task{
		ExpressionID:  "expr1",
		TaskID:        1,
		Arg1:          10,
		Arg2:          5,
		Operation:     "+",
		OperationTime: 0,
	}, nil).Once()

	mockAdapter.On("ResultTask", "expr1", 1, 15.0).Return("ok", nil).Once()

	mockAdapter.On("GetTask").Return(models.Task{}, errors.New("test timeout")).Maybe()

	service := NewAgentService(mockAdapter)

	ctx, _ = logger.New(ctx)
	go service.Work(ctx, 1)

	<-ctx.Done()

	mockAdapter.AssertExpectations(t)
}

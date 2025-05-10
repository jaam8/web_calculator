package utils

import (
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) CreateTask(arg1, arg2 float64, oper string, ExprID int) models.Task {
	args := m.Called(arg1, arg2, oper, ExprID)
	return args.Get(0).(models.Task)
}

func (m *MockTaskManager) AddResult(result models.Result) {
	m.Called(result)
}

func (m *MockTaskManager) GetResult() models.Result {
	args := m.Called()
	return args.Get(0).(models.Result)
}

type MockExpressionManager struct {
	mock.Mock
}

func (m *MockExpressionManager) CreateExpression() (int, error) {
	return 0, nil
}

func (m *MockExpressionManager) GetTaskManager(expressionID int) (types.TaskManager, bool) {
	return nil, false
}

func (m *MockExpressionManager) GetTasks() chan models.Task {
	return nil
}

func (m *MockExpressionManager) GetExpression(id int) (*models.Expression, bool) {
	return nil, false
}

func (m *MockExpressionManager) GetExpressions() []*models.Expression {
	return nil
}

func (m *MockExpressionManager) GetExpressionByID(id int) (*models.Expression, bool) {
	return nil, false
}

func (m *MockExpressionManager) AddTask(task models.Task) {
	m.Called(task)
}

func (m *MockExpressionManager) ExpressionDone(exprID int, res float64) {
	m.Called(exprID, res)
}

func (m *MockExpressionManager) ExpressionError(exprID int) {
	m.Called(exprID)
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name        string
		rpn         []string
		expectTasks []models.Task
		wantStatus  string
		wantResult  float64
	}{
		{
			name: "valid expression",
			rpn:  []string{"3", "4", "+"},
			expectTasks: []models.Task{
				{
					ExpressionID: 1,
					TaskID:       1,
					Arg1:         3,
					Arg2:         4,
					Operation:    "+",
				},
			},
			wantStatus: "done",
			wantResult: 7,
		},
		{
			name:        "invalid expression",
			rpn:         []string{"3", "*"},
			expectTasks: nil,
			wantStatus:  "invalid expression",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTM := new(MockTaskManager)
			mockEM := new(MockExpressionManager)
			exprID := 1

			for _, task := range tt.expectTasks {
				mockTM.On("CreateTask", task.Arg1, task.Arg2, task.Operation, exprID).Return(task)
				mockEM.On("AddTask", task).Return()
				mockTM.On("GetResult").Return(
					models.Result{
						ExpressionID: exprID,
						TaskID:       task.TaskID,
						Result:       tt.wantResult,
					},
				)
			}
			if tt.wantStatus == "done" {
				mockEM.On("ExpressionDone", exprID, tt.wantResult).Return()
			}
			if tt.wantStatus == "invalid expression" {
				mockEM.On("ExpressionError", exprID).Return()
			}

			Process(tt.rpn, mockTM, mockEM, exprID)

			mockTM.AssertExpectations(t)
			mockEM.AssertExpectations(t)
		})
	}
}

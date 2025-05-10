package utils

import (
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTaskManager_CreateTask(t *testing.T) {
	tm := NewTaskManager(durations)
	task1 := tm.CreateTask(1, 2, "+", 1)
	task2 := tm.CreateTask(1, 2, "+", 2)
	require.NotEqual(t, task1, task2)
}

func TestTaskManager_GetResult(t *testing.T) {
	tm := NewTaskManager(durations)
	task := tm.CreateTask(1, 2, "+", 1)
	excepted := models.Result{
		ExpressionID: task.ExpressionID,
		TaskID:       task.TaskID,
		Result:       3,
	}
	tm.AddResult(excepted)
	got := tm.GetResult()
	require.Equal(t, excepted, got)
}

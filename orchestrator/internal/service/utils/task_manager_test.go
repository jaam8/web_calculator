package utils

import (
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestNewTaskManager(t *testing.T) {
	tm := NewTaskManager(durations)
	require.NotNil(t, tm)
	require.Equal(t, durations, tm.durations)
	require.NotNil(t, tm.resultCh)
	require.Equal(t, 0, tm.Counter)
}

func TestTaskManager_CreateTask(t *testing.T) {
	tm := NewTaskManager(durations)
	exprID := uuid.New()

	t.Run("Basic_task_creation", func(t *testing.T) {
		task := tm.CreateTask(1.5, 2.5, "+", exprID)
		require.Equal(t, exprID, task.ExpressionID)
		require.Equal(t, 1, task.TaskID)
		require.Equal(t, 1.5, task.Arg1)
		require.Equal(t, 2.5, task.Arg2)
		require.Equal(t, "+", task.Operation)
		require.Equal(t, time.Millisecond*100, task.OperationTime)
	})

	t.Run("Sequential_task_IDs", func(t *testing.T) {
		task1 := tm.CreateTask(1, 2, "+", uuid.New())
		task2 := tm.CreateTask(1, 2, "+", uuid.New())
		require.Equal(t, task1.TaskID+1, task2.TaskID)
	})

	t.Run("Different_operations", func(t *testing.T) {
		operations := []string{"+", "-", "*", "/"}
		for i, op := range operations {
			task := tm.CreateTask(1, 2, op, uuid.New())
			require.Equal(t, op, task.Operation)
			require.Equal(t, i+4, task.TaskID) // Changed from i+3 to i+4
		}
	})

	t.Run("Counter_increment", func(t *testing.T) {
		initialCounter := tm.Counter
		tm.CreateTask(1, 2, "+", uuid.New())
		require.Equal(t, initialCounter+1, tm.Counter)
	})
}

func TestTaskManager_ResultChannel(t *testing.T) {
	tm := NewTaskManager(durations)
	exprID := uuid.New()
	task := tm.CreateTask(1, 2, "+", exprID)

	t.Run("Add and get result", func(t *testing.T) {
		expected := models.Result{
			ExpressionID: exprID,
			TaskID:       task.TaskID,
			Result:       3.0,
		}
		tm.AddResult(expected)
		got := tm.GetResult()
		require.Equal(t, expected, got)
	})

	t.Run("Channel blocking", func(t *testing.T) {
		done := make(chan bool)
		go func() {
			result := models.Result{
				ExpressionID: exprID,
				TaskID:       task.TaskID,
				Result:       3.0,
			}
			tm.AddResult(result)
			done <- true
		}()

		select {
		case <-done:
			// Success: result was added and channel didn't block
		case <-time.After(time.Second):
			t.Fatal("Result channel blocked")
		}
	})
}

func TestTaskManager_ConcurrentAccess(t *testing.T) {
	tm := NewTaskManager(durations)
	var wg sync.WaitGroup
	numGoroutines := 100

	t.Run("Concurrent task creation", func(t *testing.T) {
		taskIDs := make(chan int, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				task := tm.CreateTask(1, 2, "+", uuid.New())
				taskIDs <- task.TaskID
			}()
		}

		wg.Wait()
		close(taskIDs)

		// Check if all task IDs are unique
		seenIDs := make(map[int]bool)
		for id := range taskIDs {
			require.False(t, seenIDs[id], "Duplicate task ID found: %d", id)
			seenIDs[id] = true
		}
		require.Equal(t, numGoroutines, len(seenIDs))
	})

	t.Run("Concurrent result handling", func(t *testing.T) {
		results := make(chan models.Result, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				result := models.Result{
					ExpressionID: uuid.New(),
					TaskID:       i,
					Result:       float64(i),
				}
				tm.AddResult(result)
				results <- tm.GetResult()
			}(i)
		}

		wg.Wait()
		close(results)

		// Verify all results were processed
		count := 0
		for range results {
			count++
		}
		require.Equal(t, numGoroutines, count)
	})
}

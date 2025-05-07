package agent

import (
	"bytes"
	"github.com/jaam8/web_calculator/internal/config"
	"github.com/jaam8/web_calculator/orchestrator"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/task" {
			t.Fatalf("Expected to request '/internal/task', got: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"TaskID": 1, "Arg1": 3, "Arg2": 4, "Operation": "+"}`))
	}))
	defer server.Close()

	config.Configs.RequestURL = server.URL
	config.Configs.Port = ""

	task, status, err := GetTask()
	if err != nil {
		t.Fatalf("GetTasks returned an error: %v", err)
	}
	if status != 200 {
		t.Errorf("Expected status code 200, got %d", status)
	}
	if task.TaskID == 0 {
		t.Errorf("Expected a valid task TaskID, got %d", task.TaskID)
	}
}

func TestPostResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/task" {
			t.Fatalf("Expected to request '/internal/task', got: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		expectedBody := `{"id":1,"result":7}`
		if !bytes.Equal(body, []byte(expectedBody)) {
			t.Fatalf("Expected body %s, got %s", expectedBody, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config.Configs.RequestURL = server.URL
	config.Configs.Port = ""

	result := orchestrator.Result{TaskID: 1, Result: 7.0}
	err := PostResult(result)
	if err != nil {
		t.Fatalf("PostResult returned an error: %v", err)
	}
}

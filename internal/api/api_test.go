package api

import (
	"github.com/jaam8/web_calculator/internal/orchestrator"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"3+4*2/(1-5)"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CalculateHandler(c)
	if err != nil {
		t.Fatalf("CalculateHandler returned an error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"id":`) {
		t.Errorf("Expected response body to contain '\"id\":', got %s", rec.Body.String())
	}
}

func TestExpressionsHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ExpressionsHandler(c)
	if err != nil {
		t.Fatalf("ExpressionsHandler returned an error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"expressions":`) {
		t.Errorf("Expected response body to contain '\"expressions\":', got %s", rec.Body.String())
	}
}

func TestExpressionByIDHandler(t *testing.T) {
	expressionID, _ := expressionManager.CreateExpression()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(expressionID))

	err := ExpressionByIDHandler(c)
	if err != nil {
		t.Fatalf("ExpressionByIDHandler returned an error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"id":`+strconv.Itoa(expressionID)) {
		t.Errorf("Expected response body to contain '\"id\":%d', got %s", expressionID, rec.Body.String())
	}
}

func TestPostTaskHandler(t *testing.T) {
	task := orchestrator.Task{ID: 1, Arg1: 3, Arg2: 4, Operation: "+"}
	taskManager.AddTask(task)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/internal/task", strings.NewReader(`{"id":1,"result":7}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := PostTaskHandler(c)
	if err != nil {
		t.Fatalf("PostTaskHandler returned an error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"result":7`) {
		t.Errorf("Expected response body to contain '\"result\":7', got %s", rec.Body.String())
	}
}

func TestGetTaskHandler(t *testing.T) {
	task := orchestrator.Task{ID: 1, Arg1: 3, Arg2: 4, Operation: "+"}
	taskManager.AddTask(task)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := GetTaskHandler(c)
	if err != nil {
		t.Fatalf("GetTaskHandler returned an error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"id":1`) {
		t.Errorf("Expected response body to contain '\"id\":1', got %s", rec.Body.String())
	}
}

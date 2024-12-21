package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateSuccess(t *testing.T) {
	testCases := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{"simple", "1+1", 2},
		{"priority", "(2+2)*2", 8},
		{"multiplication", "2+2*2", 6},
		{"division", "1/2", 0.5},
	}

	for _, testCase := range testCases {
		requestData := map[string]string{"expression": testCase.expression}
		requestBytes, err := json.Marshal(requestData)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewReader(requestBytes))
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		CalculateHandler(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]float64
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatal("Failed to parse JSON response:", err)
		}
		result, exists := response["result"]
		if !exists || result != testCase.expectedResult {
			t.Errorf("Expected result %.2f, got %.2f", testCase.expectedResult, result)
		}
	}
}

func TestCalculateFail(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{"invalid operator", "1+1*"},
		{"invalid characters", "2+2**2"},
		{"mismatched parentheses", "((2+2-*(2"},
		{"empty expression", ""},
		{"letter in expression", "1u*2"},
	}

	for _, testCase := range testCases {
		requestData := map[string]string{"expression": testCase.expression}
		requestBytes, err := json.Marshal(requestData)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewReader(requestBytes))
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		CalculateHandler(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Expected status 422, got %d", resp.StatusCode)
		}
	}
}

package orchestrator

import (
	"testing"
)

func TestValidateExpression_EmptyExpression_ReturnsError(t *testing.T) {
	err := ValidateExpression("")
	if err == nil {
		t.Errorf("Expected error for empty expression, got nil")
	}
}

func TestValidateExpression_InvalidCharacters_ReturnsError(t *testing.T) {
	err := ValidateExpression("2+2=4")
	if err == nil {
		t.Errorf("Expected error for invalid characters, got nil")
	}
}

func TestValidateExpression_UnbalancedBrackets_ReturnsError(t *testing.T) {
	err := ValidateExpression("(2+2")
	if err == nil {
		t.Errorf("Expected error for unbalanced brackets, got nil")
	}
}

func TestValidateExpression_OperatorAtStart_ReturnsError(t *testing.T) {
	err := ValidateExpression("+2+2")
	if err == nil {
		t.Errorf("Expected error for operator at start, got nil")
	}
}

func TestValidateExpression_OperatorAtEnd_ReturnsError(t *testing.T) {
	err := ValidateExpression("2+2+")
	if err == nil {
		t.Errorf("Expected error for operator at end, got nil")
	}
}

func TestValidateExpression_TwoOperatorsInARow_ReturnsError(t *testing.T) {
	err := ValidateExpression("2++2")
	if err == nil {
		t.Errorf("Expected error for two operators in a row, got nil")
	}
}

func TestValidateExpression_DivisionByZero_ReturnsError(t *testing.T) {
	err := ValidateExpression("2/0")
	if err == nil {
		t.Errorf("Expected error for division by zero, got nil")
	}
}

func TestValidateExpression_ValidExpression_ReturnsNoError(t *testing.T) {
	err := ValidateExpression("2+2")
	if err != nil {
		t.Errorf("Expected no error for valid expression, got %v", err)
	}
}

func TestRPN_ValidExpression_ReturnsRPN(t *testing.T) {
	rpn, err := RPN("3+4*2/(1-5)")
	expected := []string{"3", "4", "2", "*", "1", "5", "-", "/", "+"}
	if err != nil {
		t.Errorf("Expected no error for valid expression, got %v", err)
	}
	if !equal(rpn, expected) {
		t.Errorf("Expected %v, got %v", expected, rpn)
	}
}

func TestRPN_InvalidExpression_ReturnsError(t *testing.T) {
	_, err := RPN("3+4*2/(1-5)^2^3)")
	if err == nil {
		t.Errorf("Expected error for invalid expression, got nil")
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

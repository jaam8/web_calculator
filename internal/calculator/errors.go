package calculator

import "errors"

var (
	ErrInvalidExpression = errors.New("Expression is not valid")
	ErrDivisionByZero    = errors.New("Division by zero")
)

package errors

import (
	"errors"
)

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrExpressionNotFound  = errors.New("expression not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidExpression   = errors.New("invalid expression")
	ErrDivideByZero        = errors.New("division by zero")
)

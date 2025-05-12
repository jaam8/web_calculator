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
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrEmptyLogin          = errors.New("empty login")
	ErrEmptyPassword       = errors.New("empty password")
	ErrWrongPassword       = errors.New("wrong password")
)

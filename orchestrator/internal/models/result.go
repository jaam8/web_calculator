package models

import "github.com/google/uuid"

type Result struct {
	ExpressionID uuid.UUID
	TaskID       int     `json:"id"`
	Result       float64 `json:"result"`
}

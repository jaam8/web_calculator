package models

import (
	"github.com/google/uuid"
	"time"
)

type Task struct {
	ExpressionID  uuid.UUID
	TaskID        int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

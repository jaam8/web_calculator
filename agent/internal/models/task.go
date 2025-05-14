package models

import "time"

type Task struct {
	ExpressionID  string
	TaskID        int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

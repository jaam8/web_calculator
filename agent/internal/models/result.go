package models

type Result struct {
	ExpressionID string
	TaskID       int     `json:"id"`
	Result       float64 `json:"result"`
}

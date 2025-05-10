package models

type Result struct {
	ExpressionID int
	TaskID       int     `json:"id"`
	Result       float64 `json:"result"`
}

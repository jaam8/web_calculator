package models

type Expression struct {
	ExpressionID int      `json:"id"`
	Status       string   `json:"status"`
	Result       *float64 `json:"result,omitempty"`
}

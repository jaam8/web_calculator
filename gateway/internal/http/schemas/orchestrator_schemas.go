package schemas

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type Expression struct {
	Id     int      `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
}

var (
	InternalServerError   = "internal server error"
	ExpressionNotFound    = "expression not found"
	CannotParseId         = "cannot parse id"
	CannotParseExpression = "cannot parse expression"
)

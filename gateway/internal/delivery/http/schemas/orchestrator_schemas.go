package schemas

type CalculateRequest struct {
	Expression string `json:"expression" example:"40+2"`
}

type CalculateResponse struct {
	Id int `json:"id" example:"1"`
}

type Expression struct {
	Id     int      `json:"id" example:"1"`
	Status string   `json:"status" example:"done"`
	Result *float64 `json:"result,omitempty" example:"42.0"`
}

type ExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}

type ExpressionByIdRequest struct {
	Id int64 `json:"id" example:"1"`
}

type ExpressionByIdResponse struct {
	Expression
}

type InternalServerError struct {
	Error string `json:"error" example:"internal server error"`
}

type ExpressionNotFound struct {
	Error string `json:"error" example:"expression not found"`
}

type CannotParseId struct {
	Error string `json:"error" example:"cannot parse id"`
}

type CannotParseExpression struct {
	Error string `json:"error" example:"cannot parse expression"`
}

var (
	InternalServerErrorMsg   = InternalServerError{Error: "internal server error"}
	ExpressionNotFoundMsg    = ExpressionNotFound{Error: "expression not found"}
	CannotParseIdMsg         = CannotParseId{Error: "cannot parse id"}
	CannotParseExpressionMsg = CannotParseExpression{Error: "cannot parse expression"}
)

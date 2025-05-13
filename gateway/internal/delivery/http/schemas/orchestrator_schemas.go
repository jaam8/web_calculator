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

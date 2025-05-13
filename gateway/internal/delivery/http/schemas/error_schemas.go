package schemas

// region auth_service

type CannotParseRequest struct {
	Error string `json:"error" example:"cannot parse expression"`
}
type TokenExpiredOrInvalid struct {
	Error string `json:"error" example:"token expired or invalid"`
}

type TokenExpired struct {
	Error string `json:"error" example:"token expired"`
}

type EmptyLogin struct {
	Error string `json:"error" example:"empty login"`
}

type EmptyPassword struct {
	Error string `json:"error" example:"empty password"`
}

type WrongCredentials struct {
	Error string `json:"error" example:"wrong credentials, user not found"`
}

var (
	CannotParseRequestMsg    = CannotParseRequest{Error: "cannot parse request"}
	EmptyLoginMsg            = EmptyLogin{Error: "empty login"}
	EmptyPasswordMsg         = EmptyPassword{Error: "empty password"}
	WrongCredentialsMsg      = WrongCredentials{Error: "wrong credentials, user not found"}
	TokenExpiredOrInvalidMsg = TokenExpiredOrInvalid{Error: "token expired or invalid"}
	TokenExpiredMsg          = TokenExpired{Error: "token expired"}
)

// endregion auth_service

// region general

type InternalServerError struct {
	Error string `json:"error" example:"internal server error"`
}

var (
	InternalServerErrorMsg = InternalServerError{Error: "internal server error"}
)

// endregion general

// region orchestrator

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
	ExpressionNotFoundMsg    = ExpressionNotFound{Error: "expression not found"}
	CannotParseIdMsg         = CannotParseId{Error: "cannot parse id"}
	CannotParseExpressionMsg = CannotParseExpression{Error: "cannot parse expression"}
)

// endregion orchestrator

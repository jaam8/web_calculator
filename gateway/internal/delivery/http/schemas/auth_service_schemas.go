package schemas

type LoginRequest struct {
	Login    string `json:"login" example:"qwerty"`
	Password string `json:"password" example:"qwerty123"`
}

type RegisterRequest struct {
	Login    string `json:"login" example:"qwerty"`
	Password string `json:"password" example:"qwerty123"`
}

type RegisterResponse struct {
	UserId string `json:"user_id" example:"0196cb7d-7d60-78cc-ac28-f9e114de51fc"`
}

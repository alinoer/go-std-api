package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User    *User  `json:"user"`
	Message string `json:"message"`
}
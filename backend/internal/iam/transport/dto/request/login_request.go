package request

// LoginRequest is the HTTP request body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

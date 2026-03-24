package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

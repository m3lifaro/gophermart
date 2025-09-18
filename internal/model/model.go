package model

type CreateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	UUID  string `json:"uuid"`
	Login string `json:"login"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type jwtKey = string

const (
	UserIDKey jwtKey = "uuid"
	LoginKey  jwtKey = "login"
	EmailKey  jwtKey = "email"
	RoleKey   jwtKey = "role"
)

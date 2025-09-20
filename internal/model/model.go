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
	ID    int32  `json:"uid" db:"id"`
	Login string `json:"login" db:"login"`
	//Email string `json:"email"`
	//Role  string `json:"role"`
}
type UserDao struct {
	Password string `json:"password" db:"password"`
	User
}
type jwtKey = string

const (
	UserIDKey jwtKey = "uid"
	LoginKey  jwtKey = "login"
	//EmailKey  jwtKey = "email"
	//RoleKey   jwtKey = "role"
)

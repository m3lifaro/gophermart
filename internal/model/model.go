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

type OrderItem struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type ExternalOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

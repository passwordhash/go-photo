package request

type AuthLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthRegister struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

package model

type RegisterParams struct {
	Email    string
	Password string
}

type RegisterInfo struct {
	UserUUID string
	Token    string
}

type TokenPayload struct {
	UserUUID string
}

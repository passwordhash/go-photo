package response

type Login struct {
	Token string `json:"token"`
}

type Register struct {
	UserUUID string `json:"user_uuid"`
	Token    string `json:"token"`
}

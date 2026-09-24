package model

// LoginInput 登录参数。
type LoginInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Password string `json:"password" v:"required|length:5,72"`
}

// RegisterInput 注册参数。
type RegisterInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Password string `json:"password" v:"required|length:6,72"`
	Nickname string `json:"nickname" v:"length:0,100"`
	PhoneNum string `json:"phone_num" v:"length:0,32"`
	Email    string `json:"email" v:"email"`
	OrgID    int64  `json:"org_id" v:"required"`
}

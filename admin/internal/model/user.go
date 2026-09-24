package model

import "time"

// User 用户数据库实体。
type User struct {
	ID        int64      `json:"id"`
	UID       int64      `json:"uid"`
	Username  string     `json:"username"`
	Nickname  string     `json:"nickname"`
	PhoneNum  string     `json:"phone_num"`
	Email     string     `json:"email"`
	Password  string     `json:"-" orm:"password"`
	OrgID     int64      `json:"org_id"`
	Status    int        `json:"-" orm:"status"`
	Extra     *string    `json:"extra"`
	Version   int        `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// UserView 用户接口视图。
type UserView struct {
	ID        int64      `json:"id"`
	UID       int64      `json:"uid"`
	Username  string     `json:"username"`
	Nickname  string     `json:"nickname"`
	PhoneNum  string     `json:"phone_num"`
	Email     string     `json:"email"`
	OrgID     int64      `json:"org_id"`
	Status    string     `json:"status"`
	Version   int        `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// UserCreateInput 管理员创建用户参数。
type UserCreateInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Nickname string `json:"nickname" v:"length:0,100"`
	PhoneNum string `json:"phone_num" v:"length:0,32"`
	Email    string `json:"email" v:"email"`
	Password string `json:"password" v:"length:0,72"`
	OrgID    int64  `json:"org_id" v:"required"`
	Status   int    `json:"status"`
}

// UserCreateResult 管理员新建用户结果，初始密码只在创建成功时返回一次。
type UserCreateResult struct {
	User            *UserView `json:"user"`
	InitialPassword string    `json:"initial_password"`
}

// UserUpdateInput 用户修改参数。
type UserUpdateInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Nickname string `json:"nickname" v:"length:0,100"`
	PhoneNum string `json:"phone_num" v:"length:0,32"`
	Email    string `json:"email" v:"email"`
	Password string `json:"password" v:"length:0,72"`
	Status   int    `json:"status"`
	Version  int    `json:"version" v:"required"`
}

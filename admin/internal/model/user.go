package model

import "time"

// User 全局用户数据库实体。
type User struct {
	ID           int64      `json:"id"`
	UID          int64      `json:"uid"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	Phone        string     `json:"-"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Status       int        `json:"-"`
	Extra        *string    `json:"extra"`
	Version      int        `json:"version"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type UserView struct {
	ID        int64      `json:"id"`
	UID       int64      `json:"uid"`
	Username  string     `json:"username"`
	Nickname  string     `json:"nickname"`
	PhoneNum  string     `json:"phone_num"`
	Email     string     `json:"email"`
	TenantIDs []int64    `json:"tenant_ids"`
	Status    string     `json:"status"`
	Version   int        `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserCreateInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Nickname string `json:"nickname" v:"length:0,100"`
	PhoneNum string `json:"phone_num" v:"length:0,32"`
	Email    string `json:"email" v:"email"`
	Password string `json:"password" v:"length:0,72"`
	TenantID int64  `json:"tenant_id" v:"required"`
	Status   *int   `json:"status"`
}

type UserCreateResult struct {
	User            *UserView `json:"user"`
	InitialPassword string    `json:"initial_password"`
}

type UserUpdateInput struct {
	Username string `json:"username" v:"required|length:3,64"`
	Nickname string `json:"nickname" v:"length:0,100"`
	PhoneNum string `json:"phone_num" v:"length:0,32"`
	Email    string `json:"email" v:"email"`
	Password string `json:"password" v:"length:0,72"`
	Status   int    `json:"status"`
	Version  int    `json:"version" v:"required"`
}

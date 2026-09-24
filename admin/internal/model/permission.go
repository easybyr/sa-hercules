package model

import "time"

// Permission 权限实体。
type Permission struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OrgID       int64      `json:"org_id"`
	Extra       *string    `json:"extra"`
	Version     int        `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// PermissionInput 权限新建参数。
type PermissionInput struct {
	Code        string `json:"code" v:"required|length:3,100"`
	Name        string `json:"name" v:"required|length:1,100"`
	Description string `json:"description" v:"length:0,500"`
	OrgID       int64  `json:"org_id" v:"required"`
}

// PermissionUpdateInput 权限修改参数。
type PermissionUpdateInput struct {
	Name        string `json:"name" v:"required|length:1,100"`
	Description string `json:"description" v:"length:0,500"`
	Version     int    `json:"version" v:"required"`
}

package model

import "time"

// Role 角色实体。
type Role struct {
	ID          int64      `json:"id"`
	RoleID      int64      `json:"role_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OrgID       int64      `json:"org_id"`
	Extra       *string    `json:"extra"`
	Version     int        `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// RoleInput 角色新建参数。
type RoleInput struct {
	Name            string   `json:"name" v:"required|length:1,100"`
	Description     string   `json:"description" v:"length:0,500"`
	OrgID           int64    `json:"org_id" v:"required"`
	PermissionCodes []string `json:"permission_codes"`
}

// RoleUpdateInput 角色修改参数。
type RoleUpdateInput struct {
	Name            string   `json:"name" v:"required|length:1,100"`
	Description     string   `json:"description" v:"length:0,500"`
	Version         int      `json:"version" v:"required"`
	PermissionCodes []string `json:"permission_codes"`
}

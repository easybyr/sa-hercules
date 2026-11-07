package model

import "time"

type Role struct {
	ID          int64      `json:"id"`
	RoleID      int64      `json:"role_id"` // API compatibility alias for ID.
	TenantID    int64      `json:"tenant_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      int        `json:"-"`
	IsSystem    int        `json:"is_system"`
	Extra       *string    `json:"extra,omitempty"`
	Version     int        `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type RoleInput struct {
	Name            string   `json:"name" v:"required|length:1,128"`
	Description     string   `json:"description" v:"length:0,500"`
	TenantID        int64    `json:"tenant_id" v:"required"`
	PermissionCodes []string `json:"permission_codes"`
}

type RoleUpdateInput struct {
	Name            string   `json:"name" v:"required|length:1,128"`
	Description     string   `json:"description" v:"length:0,500"`
	Version         int      `json:"version" v:"required"`
	PermissionCodes []string `json:"permission_codes"`
}

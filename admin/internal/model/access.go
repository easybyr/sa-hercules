package model

import "time"

// UserRole 用户角色实体。
type UserRole struct {
	ID        int64      `json:"id"`
	UID       int64      `json:"uid"`
	RoleID    int64      `json:"role_id"`
	Extra     *string    `json:"extra"`
	Version   int        `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// RolePermission 角色权限实体。
type RolePermission struct {
	ID             int64      `json:"id"`
	RoleID         int64      `json:"role_id"`
	PermissionCode string     `json:"permission_code"`
	Extra          *string    `json:"extra"`
	Version        int        `json:"version"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// RoleAccessView 角色及其当前组织内的权限列表。
type RoleAccessView struct {
	Role        *Role         `json:"role"`
	Permissions []*Permission `json:"permissions"`
}

// IDListInput 角色ID列表参数。
type IDListInput struct {
	RoleIDs []int64 `json:"role_ids"`
}

// CodeListInput 权限编码列表参数。
type CodeListInput struct {
	PermissionCodes []string `json:"permission_codes"`
}

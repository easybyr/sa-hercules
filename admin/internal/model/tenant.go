package model

import "time"

// Tenant 租户数据库实体。
type Tenant struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	ParentCode  *string    `json:"parent_code"`
	Name        string     `json:"name"`
	Status      int        `json:"-"`
	Description string     `json:"description"`
	Extra       *string    `json:"extra"`
	Version     int        `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// TenantView 租户接口视图。
type TenantView struct {
	ID          int64         `json:"id"`
	TenantID    int64         `json:"tenant_id"`
	Code        string        `json:"code"`
	ParentCode  *string       `json:"parent_code"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Description string        `json:"description"`
	Version     int           `json:"version"`
	CreatedAt   *time.Time    `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at"`
	Children    []*TenantView `json:"children,omitempty"`
}

type TenantInput struct {
	Name        string  `json:"name" v:"required|length:1,128"`
	ParentCode  *string `json:"parent_code"`
	Description string  `json:"description" v:"length:0,500"`
	Status      *int    `json:"status"`
}

type TenantUpdateInput struct {
	Name        string  `json:"name" v:"required|length:1,128"`
	ParentCode  *string `json:"parent_code"`
	Description string  `json:"description" v:"length:0,500"`
	Status      int     `json:"status" v:"required"`
	Version     int     `json:"version" v:"required"`
}

type TenantAssignmentInput struct {
	TenantID int64 `json:"tenant_id" v:"required"`
}

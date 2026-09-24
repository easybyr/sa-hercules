package model

import "time"

// Organization 组织数据库实体。
type Organization struct {
	ID          int64      `json:"id"`
	OrgID       int64      `json:"org_id"`
	Name        string     `json:"name"`
	ParentID    *int64     `json:"parent_id"`
	Description string     `json:"description"`
	Status      int        `json:"-" orm:"status"`
	Extra       *string    `json:"extra"`
	Version     int        `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// OrganizationView 组织接口视图。
type OrganizationView struct {
	ID          int64               `json:"id"`
	OrgID       int64               `json:"org_id"`
	Name        string              `json:"name"`
	ParentID    *int64              `json:"parent_id"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Version     int                 `json:"version"`
	CreatedAt   *time.Time          `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at"`
	Children    []*OrganizationView `json:"children,omitempty"`
}

// OrganizationInput 组织新建参数。
type OrganizationInput struct {
	Name        string `json:"name" v:"required|length:1,100"`
	ParentID    *int64 `json:"parent_id"`
	Description string `json:"description" v:"length:0,500"`
	Status      int    `json:"status"`
}

// OrganizationUpdateInput 组织修改参数。
type OrganizationUpdateInput struct {
	Name        string `json:"name" v:"required|length:1,100"`
	ParentID    *int64 `json:"parent_id"`
	Description string `json:"description" v:"length:0,500"`
	Status      int    `json:"status" v:"required"`
	Version     int    `json:"version" v:"required"`
}

// OrganizationAssignmentInput 用户组织分配参数。
type OrganizationAssignmentInput struct {
	OrgID int64 `json:"org_id" v:"required"`
}

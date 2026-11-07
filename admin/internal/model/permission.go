package model

import (
	"time"

	"github.com/example/sa-hercules/admin/internal/consts"
)

type Permission struct {
	ID          int64                     `json:"id"`
	TenantID    int64                     `json:"tenant_id"`
	Code        string                    `json:"code"`
	Name        string                    `json:"name"`
	NodeType    consts.PermissionNodeType `json:"node_type"`
	ParentID    *int64                    `json:"parent_id"`
	Description string                    `json:"description"`
	Status      int                       `json:"-"`
	APIs        []string                  `json:"apis"`
	Extra       *string                   `json:"extra,omitempty"`
	Version     int                       `json:"version"`
	CreatedAt   *time.Time                `json:"created_at"`
	UpdatedAt   *time.Time                `json:"updated_at"`
}

type PermissionInput struct {
	Code        string                    `json:"code" v:"required|length:3,128"`
	NodeType    consts.PermissionNodeType `json:"node_type" v:"required"`
	ParentID    *int64                    `json:"parent_id"`
	APIs        []string                  `json:"apis"`
	Name        string                    `json:"name" v:"required|length:1,128"`
	Description string                    `json:"description" v:"length:0,255"`
	TenantID    int64                     `json:"tenant_id" v:"required"`
}

type PermissionUpdateInput struct {
	Code        string                    `json:"code" v:"length:3,128"`
	NodeType    consts.PermissionNodeType `json:"node_type"`
	ParentID    *int64                    `json:"parent_id"`
	APIs        []string                  `json:"apis"`
	Name        string                    `json:"name" v:"required|length:1,128"`
	Description string                    `json:"description" v:"length:0,255"`
	Version     int                       `json:"version" v:"required"`
}

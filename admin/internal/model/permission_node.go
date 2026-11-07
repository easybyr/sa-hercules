package model

import (
	"time"

	"github.com/example/sa-hercules/admin/internal/consts"
)

// PermissionNode 权限节点关联的 API 路径集合。
type PermissionNode struct {
	ID             int64                     `json:"id"`
	PermissionCode string                    `json:"permission_code"`
	NodeType       consts.PermissionNodeType `json:"node_type"`
	APIs           string                    `json:"-" db:"apis"`
	APIsHash       []byte                    `json:"-" db:"apis_hash"`
	Extra          *string                   `json:"extra"`
	Version        int                       `json:"version"`
	CreatedAt      *time.Time                `json:"created_at"`
	UpdatedAt      *time.Time                `json:"updated_at"`
}

// APIPermission 已注册的 HTTP API 及其权限编码。
type APIPermission struct {
	ID             int64      `json:"id"`
	Method         string     `json:"method"`
	Path           string     `json:"path"`
	Service        string     `json:"service"`
	PermissionCode string     `json:"permission_code"`
	PathType       string     `json:"path_type,omitempty"`
	Description    string     `json:"description,omitempty"`
	Status         int        `json:"status,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

package model

import "time"

type TenantMember struct {
	ID        int64      `json:"id"`
	TenantID  int64      `json:"tenant_id"`
	UserID    int64      `json:"user_id"`
	Status    int        `json:"status"`
	JoinedAt  *time.Time `json:"joined_at"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserRole struct {
	ID             int64      `json:"id"`
	TenantID       int64      `json:"tenant_id"`
	TenantMemberID int64      `json:"tenant_member_id"`
	UserID         int64      `json:"user_id"`
	RoleID         int64      `json:"role_id"`
	CreatedAt      *time.Time `json:"created_at"`
}

type RolePermission struct {
	ID           int64      `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	RoleID       int64      `json:"role_id"`
	PermissionID int64      `json:"permission_id"`
	CreatedAt    *time.Time `json:"created_at"`
}

type PermissionOperationLog struct {
	ID          int64      `json:"id" db:"id"`
	TenantID    int64      `json:"tenant_id" db:"tenant_id"`
	OperatorUID int64      `json:"operator_uid" db:"operator_uid"`
	TargetType  string     `json:"target_type" db:"target_type"`
	TargetID    int64      `json:"target_id" db:"target_id"`
	Operation   string     `json:"operation" db:"operation"`
	BeforeData  string     `json:"before_data" db:"before_data"`
	AfterData   string     `json:"after_data" db:"after_data"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
}

type RoleAccessView struct {
	Role        *Role         `json:"role"`
	Permissions []*Permission `json:"permissions"`
}

type IDListInput struct {
	RoleIDs  []int64 `json:"role_ids"`
	TenantID int64   `json:"tenant_id"`
}

type CodeListInput struct {
	PermissionCodes []string `json:"permission_codes"`
	TenantID        int64    `json:"tenant_id"`
}

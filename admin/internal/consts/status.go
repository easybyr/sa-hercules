package consts

// Status 资源状态。
type Status int

const (
	// StatusNormal 正常。
	StatusNormal Status = 1
	// StatusDisabled 禁用。
	StatusDisabled Status = 0
)

const (
	// SuperAdminUID 超级管理员UID。
	SuperAdminUID int64 = 1000000000
	// SuperAdminTenantCode 是系统管理租户的稳定业务编码。
	SuperAdminTenantCode = "10000000"
	// SuperAdminRoleCode 超级管理员角色编码。
	SuperAdminRoleCode = "super_admin"
	// TenantAdminRoleCode 租户管理员角色编码。
	TenantAdminRoleCode = "tenant_admin"
)

// String 返回用于接口输出的状态名称。
func (s Status) String() string {
	switch s {
	case StatusNormal:
		return "normal"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// Valid 校验状态值是否合法。
func (s Status) Valid() bool {
	return s == StatusNormal || s == StatusDisabled
}

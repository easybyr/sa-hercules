package consts

// Status 资源状态。
type Status int

const (
	// StatusNormal 正常。
	StatusNormal Status = 1
	// StatusDisabled 禁用。
	StatusDisabled Status = 2
)

const (
	// SuperAdminUID 超级管理员UID。
	SuperAdminUID int64 = 1000000000
	// SuperAdminOrgID 超级管理员组织ID。
	SuperAdminOrgID int64 = 10000000
	// SuperAdminRoleID 超级管理员角色ID。
	SuperAdminRoleID int64 = 10000000
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

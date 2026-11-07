package consts

import "strings"

// PermissionNodeType 权限节点类型。
type PermissionNodeType string

const (
	// PermissionNodeTypeMenu 菜单权限。
	PermissionNodeTypeMenu PermissionNodeType = "MENU"
	// PermissionNodeTypeButton 按钮权限。
	PermissionNodeTypeButton PermissionNodeType = "BUTTON"
)

// Valid 判断权限节点类型是否合法。
func (t PermissionNodeType) Valid() bool {
	return t == PermissionNodeTypeMenu || t == PermissionNodeTypeButton || t == "API" || t == "DATA"
}

// PermissionNodeTypeForCode 根据权限编码前缀解析节点类型。
func PermissionNodeTypeForCode(code string) PermissionNodeType {
	for _, nodeType := range []PermissionNodeType{"MENU", "BUTTON", "API", "DATA"} {
		if strings.HasPrefix(code, strings.ToLower(string(nodeType))+".") {
			return nodeType
		}
	}
	return ""
}

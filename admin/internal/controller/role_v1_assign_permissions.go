package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AssignRolePermissions 分配角色权限。
func (c *Controller) AssignRolePermissions(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.assign_permission")
	if !ok {
		return
	}
	var input model.CodeListInput
	if !c.parse(r, &input) {
		return
	}
	if err := c.service.AssignRolePermissions(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("roleId").Int64(),
		input.PermissionCodes,
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

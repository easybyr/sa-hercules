package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AssignUserPermissions 分配用户直接权限。
func (c *Controller) AssignUserPermissions(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.assign_permission")
	if !ok {
		return
	}
	var input model.CodeListInput
	if !c.parse(r, &input) {
		return
	}
	if err := c.service.AssignUserPermissions(
		r.Context(), claims, r.Get("uid").Int64(), input.PermissionCodes,
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

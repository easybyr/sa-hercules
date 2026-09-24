package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AssignUserRoles 分配用户角色。
func (c *Controller) AssignUserRoles(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.assign_role")
	if !ok {
		return
	}
	var input model.IDListInput
	if !c.parse(r, &input) {
		return
	}
	if err := c.service.AssignUserRoles(
		r.Context(), claims, r.Get("uid").Int64(), input.RoleIDs,
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

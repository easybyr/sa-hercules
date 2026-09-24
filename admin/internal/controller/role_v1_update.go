package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateRole 修改角色。
func (c *Controller) UpdateRole(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.update")
	if !ok {
		return
	}
	var input model.RoleUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdateRole(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("roleId").Int64(), &input,
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

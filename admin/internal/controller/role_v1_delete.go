package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeleteRole 删除角色。
func (c *Controller) DeleteRole(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.delete")
	if !ok {
		return
	}
	if err := c.service.DeleteRole(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("roleId").Int64(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

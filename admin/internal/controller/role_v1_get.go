package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RoleDetail 查询角色详情。
func (c *Controller) RoleDetail(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.read")
	if !ok {
		return
	}
	data, err := c.service.RoleDetail(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("roleId").Int64(),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

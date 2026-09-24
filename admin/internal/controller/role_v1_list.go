package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListRoles 查询角色。
func (c *Controller) ListRoles(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.read")
	if !ok {
		return
	}
	data, err := c.service.ListRoles(
		r.Context(),
		claims,
		r.Get("org_id").Int64(),
		r.Get("q").String(),
		pageRequest(r),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

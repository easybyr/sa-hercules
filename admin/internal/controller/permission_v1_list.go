package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListPermissions 查询权限。
func (c *Controller) ListPermissions(r *ghttp.Request) {
	claims, ok := c.authorize(r, "permission.read")
	if !ok {
		return
	}
	data, err := c.service.ListPermissions(
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

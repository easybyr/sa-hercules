package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListOrgMembers 查询组织成员。
func (c *Controller) ListOrgMembers(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.read")
	if !ok {
		return
	}
	data, err := c.service.ListUsers(
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

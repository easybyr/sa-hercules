package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListTenantMembers 查询租户成员。
func (c *Controller) ListTenantMembers(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionUserRead)
	if !ok {
		return
	}
	data, err := c.service.ListUsers(
		r.Context(),
		claims,
		r.Get("tenant_id").Int64(),
		r.Get("q").String(),
		pageRequest(r),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

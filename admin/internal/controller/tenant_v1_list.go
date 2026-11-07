package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListTenants 分页查询租户。
func (c *Controller) ListTenants(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionTenantRead)
	if !ok {
		return
	}
	data, err := c.service.ListTenants(
		r.Context(),
		claims,
		r.Get("q").String(),
		pageRequest(r),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeleteTenant 删除租户。
func (c *Controller) DeleteTenant(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionTenantDelete)
	if !ok {
		return
	}
	if err := c.service.DeleteTenant(
		r.Context(), claims, r.Get("id").Int64(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

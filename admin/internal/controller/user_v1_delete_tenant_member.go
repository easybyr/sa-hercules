package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeleteTenantMember 删除租户成员。
func (c *Controller) DeleteTenantMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionUserDelete)
	if !ok {
		return
	}
	if err := c.service.DeleteUser(
		r.Context(), claims, r.Get("uid").Int64(), r.Get("tenant_id").Int64(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

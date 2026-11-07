package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// GetTenantMember 查询租户成员详情。
func (c *Controller) GetTenantMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionUserRead)
	if !ok {
		return
	}
	data, err := c.service.UserDetail(
		r.Context(), claims, r.Get("uid").Int64(), r.Get("tenant_id").Int64(),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

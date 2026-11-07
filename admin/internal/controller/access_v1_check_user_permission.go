package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CheckUserPermission 检查指定用户是否属于租户并具备权限编码。
func (c *Controller) CheckUserPermission(r *ghttp.Request) {
	allowed, err := c.service.UserHasPermission(
		r.Context(), r.Get("tenant_id").Int64(), r.Get("uid").Int64(),
		r.Get("permission_code").String(),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, allowed)
}

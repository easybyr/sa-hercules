package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeletePermission 删除权限。
func (c *Controller) DeletePermission(r *ghttp.Request) {
	claims, ok := c.authorize(r, "permission.delete")
	if !ok {
		return
	}
	if err := c.service.DeletePermission(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("code").String(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

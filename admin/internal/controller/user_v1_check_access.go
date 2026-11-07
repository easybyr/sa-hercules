package controller

import (
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CheckAccess 检查权限、角色和租户访问。
func (c *Controller) CheckAccess(r *ghttp.Request) {
	data, err := c.service.CheckAccess(
		r.Context(), middleware.Claims(r), r.Get("permission").String(),
		r.Get("role_id").Int64(), r.Get("tenant_id").Int64(),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

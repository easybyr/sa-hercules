package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListAPIPermissions 查询已注册 API，供配置按钮权限时选择。
func (c *Controller) ListAPIPermissions(r *ghttp.Request) {
	if _, ok := c.authorize(r, consts.PermissionPermissionRead); !ok {
		return
	}
	items, err := c.service.ListAPIPermissions(r.Context())
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, items)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdatePermission 修改权限。
func (c *Controller) UpdatePermission(r *ghttp.Request) {
	claims, ok := c.authorize(r, "permission.update")
	if !ok {
		return
	}
	var input model.PermissionUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdatePermission(
		r.Context(), claims, r.Get("orgId").Int64(), r.Get("code").String(), &input,
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

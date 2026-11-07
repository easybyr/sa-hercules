package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateTenant 修改租户。
func (c *Controller) UpdateTenant(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionTenantUpdate)
	if !ok {
		return
	}
	var input model.TenantUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdateTenant(
		r.Context(), claims, r.Get("id").Int64(), &input,
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

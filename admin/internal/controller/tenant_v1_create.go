package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CreateTenant 新建租户。
func (c *Controller) CreateTenant(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionTenantCreate)
	if !ok {
		return
	}
	var input model.TenantInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.CreateTenant(r.Context(), claims, &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateTenantMember 修改租户成员。
func (c *Controller) UpdateTenantMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionUserUpdate)
	if !ok {
		return
	}
	var input model.UserUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdateUser(
		r.Context(), claims, r.Get("uid").Int64(), &input,
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AssignUserTenant 分配用户租户。
func (c *Controller) AssignUserTenant(r *ghttp.Request) {
	claims, ok := c.authorize(r, consts.PermissionUserAssignTenant)
	if !ok {
		return
	}
	var input model.TenantAssignmentInput
	if !c.parse(r, &input) {
		return
	}
	if err := c.service.AssignUserTenant(
		r.Context(), claims, r.Get("uid").Int64(), input.TenantID,
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

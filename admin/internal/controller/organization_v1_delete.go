package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeleteOrganization 删除组织。
func (c *Controller) DeleteOrganization(r *ghttp.Request) {
	claims, ok := c.authorize(r, "organization.delete")
	if !ok {
		return
	}
	if err := c.service.DeleteOrganization(
		r.Context(), claims, r.Get("id").Int64(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

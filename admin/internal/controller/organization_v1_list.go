package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListOrganizations 分页查询组织。
func (c *Controller) ListOrganizations(r *ghttp.Request) {
	claims, ok := c.authorize(r, "organization.read")
	if !ok {
		return
	}
	data, err := c.service.ListOrganizations(
		r.Context(),
		claims,
		r.Get("q").String(),
		pageRequest(r),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

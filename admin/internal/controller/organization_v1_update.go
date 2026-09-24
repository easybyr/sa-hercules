package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateOrganization 修改组织。
func (c *Controller) UpdateOrganization(r *ghttp.Request) {
	claims, ok := c.authorize(r, "organization.update")
	if !ok {
		return
	}
	var input model.OrganizationUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdateOrganization(
		r.Context(), claims, r.Get("id").Int64(), &input,
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CreateOrganization 新建组织。
func (c *Controller) CreateOrganization(r *ghttp.Request) {
	claims, ok := c.authorize(r, "organization.create")
	if !ok {
		return
	}
	var input model.OrganizationInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.CreateOrganization(r.Context(), claims, &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

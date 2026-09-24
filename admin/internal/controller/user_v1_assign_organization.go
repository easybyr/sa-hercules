package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AssignUserOrganization 分配用户组织。
func (c *Controller) AssignUserOrganization(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.assign_organization")
	if !ok {
		return
	}
	var input model.OrganizationAssignmentInput
	if !c.parse(r, &input) {
		return
	}
	if err := c.service.AssignUserOrganization(
		r.Context(), claims, r.Get("uid").Int64(), input.OrgID,
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

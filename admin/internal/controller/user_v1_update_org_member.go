package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateOrgMember 修改组织成员。
func (c *Controller) UpdateOrgMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.update")
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

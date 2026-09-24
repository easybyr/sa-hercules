package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeleteOrgMember 删除组织成员。
func (c *Controller) DeleteOrgMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.delete")
	if !ok {
		return
	}
	if err := c.service.DeleteUser(
		r.Context(), claims, r.Get("uid").Int64(),
	); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, nil)
}

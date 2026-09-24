package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// GetOrgMember 查询组织成员详情。
func (c *Controller) GetOrgMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.read")
	if !ok {
		return
	}
	data, err := c.service.UserDetail(r.Context(), claims, r.Get("uid").Int64())
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SearchOrganizations 搜索注册组织。
func (c *Controller) SearchOrganizations(r *ghttp.Request) {
	data, err := c.service.SearchOrganizations(r.Context(), r.Get("q").String())
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

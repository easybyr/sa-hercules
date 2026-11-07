package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SearchTenants 搜索注册租户。
func (c *Controller) SearchTenants(r *ghttp.Request) {
	data, err := c.service.SearchTenants(r.Context(), r.Get("q").String())
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

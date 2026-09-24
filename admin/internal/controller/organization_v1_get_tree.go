package controller

import (
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OrganizationTree 获取组织树。
func (c *Controller) OrganizationTree(r *ghttp.Request) {
	if _, ok := c.authorize(r, "organization.read"); !ok {
		return
	}
	data, err := c.service.OrganizationTree(
		r.Context(),
		middleware.Claims(r),
		r.Get("q").String(),
	)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

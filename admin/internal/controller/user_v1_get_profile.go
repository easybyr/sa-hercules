package controller

import (
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Me 获取个人中心。
func (c *Controller) Me(r *ghttp.Request) {
	data, err := c.service.Me(r.Context(), middleware.Claims(r))
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

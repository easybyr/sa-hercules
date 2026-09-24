package controller

import (
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Health 服务健康检查。
func (c *Controller) Health(r *ghttp.Request) {
	if err := c.service.Health(r.Context()); err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, map[string]string{"status": "healthy"})
}

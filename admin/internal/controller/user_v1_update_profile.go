package controller

import (
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UpdateMe 修改个人信息。
func (c *Controller) UpdateMe(r *ghttp.Request) {
	var input model.UserUpdateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.UpdateMe(r.Context(), middleware.Claims(r), &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

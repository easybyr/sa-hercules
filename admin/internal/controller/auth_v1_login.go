package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Login 用户登录。
func (c *Controller) Login(r *ghttp.Request) {
	var input model.LoginInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.Login(r.Context(), &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.OK(r, data)
}

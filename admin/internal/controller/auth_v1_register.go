package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Register 用户注册。
func (c *Controller) Register(r *ghttp.Request) {
	var input model.RegisterInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.Register(r.Context(), &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

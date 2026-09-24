package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CreateRole 新建角色。
func (c *Controller) CreateRole(r *ghttp.Request) {
	claims, ok := c.authorize(r, "role.create")
	if !ok {
		return
	}
	var input model.RoleInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.CreateRole(r.Context(), claims, &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

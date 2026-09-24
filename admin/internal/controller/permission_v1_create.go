package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CreatePermission 新建权限。
func (c *Controller) CreatePermission(r *ghttp.Request) {
	claims, ok := c.authorize(r, "permission.create")
	if !ok {
		return
	}
	var input model.PermissionInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.CreatePermission(r.Context(), claims, &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CreateOrgMember 新建组织成员。
func (c *Controller) CreateOrgMember(r *ghttp.Request) {
	claims, ok := c.authorize(r, "user.create")
	if !ok {
		return
	}
	var input model.UserCreateInput
	if !c.parse(r, &input) {
		return
	}
	data, err := c.service.CreateUser(r.Context(), claims, &input)
	if err != nil {
		response.Error(r, err)
		return
	}
	response.Created(r, data)
}

package controller

import (
	"net/http"

	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/internal/service"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/example/sa-hercules/admin/pkg/token"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Controller 处理后台管理HTTP请求。
type Controller struct {
	service *service.Service
}

// New 创建控制器。
func New(svc *service.Service) *Controller {
	return &Controller{service: svc}
}

// RegisterRoutes 注册全部API路由。
func (c *Controller) RegisterRoutes(server *ghttp.Server, mw *middleware.Middleware) {
	server.Use(mw.CORS)
	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/health", c.Health)
		group.POST("/auth/login", c.Login)
		group.POST("/auth/register", c.Register)
		group.GET("/public/organizations/search", c.SearchOrganizations)

		group.Group("/", func(protected *ghttp.RouterGroup) {
			protected.Middleware(mw.Auth)
			protected.GET("/me", c.Me)
			protected.PUT("/me", c.UpdateMe)
			protected.GET("/access/check", c.CheckAccess)

			protected.GET("/organizations", c.ListOrganizations)
			protected.GET("/organizations/tree", c.OrganizationTree)
			protected.POST("/organizations", c.CreateOrganization)
			protected.PUT("/organizations/:id", c.UpdateOrganization)
			protected.DELETE("/organizations/:id", c.DeleteOrganization)

			protected.GET("/users", c.ListOrgMembers)
			protected.POST("/users", c.CreateOrgMember)
			protected.GET("/users/:uid", c.GetOrgMember)
			protected.PUT("/users/:uid", c.UpdateOrgMember)
			protected.DELETE("/users/:uid", c.DeleteOrgMember)
			protected.PUT("/users/:uid/roles", c.AssignUserRoles)
			protected.PUT("/users/:uid/permissions", c.AssignUserPermissions)
			protected.PUT("/users/:uid/organization", c.AssignUserOrganization)

			protected.GET("/permissions", c.ListPermissions)
			protected.POST("/permissions", c.CreatePermission)
			protected.PUT("/permissions/:orgId/:code", c.UpdatePermission)
			protected.DELETE("/permissions/:orgId/:code", c.DeletePermission)

			protected.GET("/roles", c.ListRoles)
			protected.POST("/roles", c.CreateRole)
			protected.GET("/roles/:orgId/:roleId", c.RoleDetail)
			protected.PUT("/roles/:orgId/:roleId", c.UpdateRole)
			protected.DELETE("/roles/:orgId/:roleId", c.DeleteRole)
			protected.PUT("/roles/:orgId/:roleId/permissions", c.AssignRolePermissions)
		})
	})
}

func (c *Controller) parse(r *ghttp.Request, target any) bool {
	if err := r.Parse(target); err != nil {
		response.Error(r, apperror.New(http.StatusBadRequest, "INVALID_ARGUMENT", err.Error()))
		return false
	}
	return true
}

func (c *Controller) authorize(r *ghttp.Request, permission string) (*token.Claims, bool) {
	claims := middleware.Claims(r)
	if claims == nil {
		response.Error(r, apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", "请先登录"))
		return nil, false
	}
	if err := c.service.RequirePermission(r.Context(), claims, permission); err != nil {
		response.Error(r, err)
		return nil, false
	}
	return claims, true
}

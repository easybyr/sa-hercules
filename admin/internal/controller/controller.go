package controller

import (
	"net/http"
	"strings"
	"unicode"

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
		group.GET("/public/tenants/search", c.SearchTenants)

		group.Group("/", func(protected *ghttp.RouterGroup) {
			protected.Middleware(mw.Auth)
			protected.GET("/me", c.Me)
			protected.PUT("/me", c.UpdateMe)
			protected.GET("/access/check", c.CheckAccess)
			protected.GET("/access/user-permission", c.CheckUserPermission)

			protected.GET("/tenants", c.ListTenants)
			protected.GET("/tenants/tree", c.TenantTree)
			protected.POST("/tenants", c.CreateTenant)
			protected.PUT("/tenants/:id", c.UpdateTenant)
			protected.DELETE("/tenants/:id", c.DeleteTenant)

			protected.GET("/users", c.ListTenantMembers)
			protected.POST("/users", c.CreateTenantMember)
			protected.GET("/users/:uid", c.GetTenantMember)
			protected.PUT("/users/:uid", c.UpdateTenantMember)
			protected.DELETE("/users/:uid", c.DeleteTenantMember)
			protected.PUT("/users/:uid/roles", c.AssignUserRoles)
			protected.PUT("/users/:uid/permissions", c.AssignUserPermissions)
			protected.PUT("/users/:uid/tenant", c.AssignUserTenant)

			protected.GET("/permissions", c.ListPermissions)
			protected.GET("/permissions/apis", c.ListAPIPermissions)
			protected.POST("/permissions", c.CreatePermission)
			protected.PUT("/permissions/:tenantId/:code", c.UpdatePermission)
			protected.DELETE("/permissions/:tenantId/:code", c.DeletePermission)

			protected.GET("/roles", c.ListRoles)
			protected.POST("/roles", c.CreateRole)
			protected.GET("/roles/:tenantId/:roleId", c.RoleDetail)
			protected.PUT("/roles/:tenantId/:roleId", c.UpdateRole)
			protected.DELETE("/roles/:tenantId/:roleId", c.DeleteRole)
			protected.PUT("/roles/:tenantId/:roleId/permissions", c.AssignRolePermissions)
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
	if r.Router != nil {
		registeredPermission, err := c.service.PermissionForAPI(
			r.Context(), r.Method, normalizeAPIRoutePath(r.Router.Uri),
		)
		if err != nil {
			response.Error(r, err)
			return nil, false
		}
		if registeredPermission != "" {
			permission = registeredPermission
		}
	}
	if err := c.service.RequirePermission(r.Context(), claims, permission); err != nil {
		response.Error(r, err)
		return nil, false
	}
	return claims, true
}

func normalizeAPIRoutePath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[index] = "{" + camelToSnake(strings.TrimPrefix(segment, ":")) + "}"
		} else if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			segments[index] = "{" + camelToSnake(name) + "}"
		}
	}
	return "/" + strings.Join(segments, "/")
}

func camelToSnake(value string) string {
	var result strings.Builder
	for index, char := range value {
		if unicode.IsUpper(char) {
			if index > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(char))
			continue
		}
		result.WriteRune(char)
	}
	return result.String()
}

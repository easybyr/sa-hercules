package bootstrap

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/gogf/gf/v2/net/ghttp"
)

// defaultAPIBindings 为现有路由指定初始权限；实际 API 清单从 GoFrame 路由表发现。
var defaultAPIBindings = []*model.APIPermission{
	{Method: "GET", Path: "/api/tenants", Service: "tenant", PermissionCode: consts.PermissionTenantRead},
	{Method: "GET", Path: "/api/tenants/tree", Service: "tenant", PermissionCode: consts.PermissionTenantRead},
	{Method: "POST", Path: "/api/tenants", Service: "tenant", PermissionCode: consts.PermissionTenantCreate},
	{Method: "PUT", Path: "/api/tenants/{id}", Service: "tenant", PermissionCode: consts.PermissionTenantUpdate},
	{Method: "DELETE", Path: "/api/tenants/{id}", Service: "tenant", PermissionCode: consts.PermissionTenantDelete},
	{Method: "GET", Path: "/api/users", Service: "user", PermissionCode: consts.PermissionUserRead},
	{Method: "POST", Path: "/api/users", Service: "user", PermissionCode: consts.PermissionUserCreate},
	{Method: "GET", Path: "/api/users/{uid}", Service: "user", PermissionCode: consts.PermissionUserRead},
	{Method: "PUT", Path: "/api/users/{uid}", Service: "user", PermissionCode: consts.PermissionUserUpdate},
	{Method: "DELETE", Path: "/api/users/{uid}", Service: "user", PermissionCode: consts.PermissionUserDelete},
	{Method: "PUT", Path: "/api/users/{uid}/roles", Service: "user", PermissionCode: consts.PermissionUserAssignRole},
	{Method: "PUT", Path: "/api/users/{uid}/permissions", Service: "user", PermissionCode: consts.PermissionUserAssignPermission},
	{Method: "PUT", Path: "/api/users/{uid}/tenant", Service: "user", PermissionCode: consts.PermissionUserAssignTenant},
	{Method: "GET", Path: "/api/permissions", Service: "permission", PermissionCode: consts.PermissionPermissionRead},
	{Method: "GET", Path: "/api/permissions/apis", Service: "permission", PermissionCode: consts.PermissionPermissionRead},
	{Method: "POST", Path: "/api/permissions", Service: "permission", PermissionCode: consts.PermissionPermissionCreate},
	{Method: "PUT", Path: "/api/permissions/{tenant_id}/{code}", Service: "permission", PermissionCode: consts.PermissionPermissionUpdate},
	{Method: "DELETE", Path: "/api/permissions/{tenant_id}/{code}", Service: "permission", PermissionCode: consts.PermissionPermissionDelete},
	{Method: "GET", Path: "/api/roles", Service: "role", PermissionCode: consts.PermissionRoleRead},
	{Method: "POST", Path: "/api/roles", Service: "role", PermissionCode: consts.PermissionRoleCreate},
	{Method: "GET", Path: "/api/roles/{tenant_id}/{role_id}", Service: "role", PermissionCode: consts.PermissionRoleRead},
	{Method: "PUT", Path: "/api/roles/{tenant_id}/{role_id}", Service: "role", PermissionCode: consts.PermissionRoleUpdate},
	{Method: "DELETE", Path: "/api/roles/{tenant_id}/{role_id}", Service: "role", PermissionCode: consts.PermissionRoleDelete},
	{Method: "PUT", Path: "/api/roles/{tenant_id}/{role_id}/permissions", Service: "role", PermissionCode: consts.PermissionRoleAssignPermission},
}

// DiscoverRegisteredAPIs 从已经物化的服务路由表构建 API 目录。
func DiscoverRegisteredAPIs(routes []ghttp.RouterItem) []*model.APIPermission {
	defaults := make(map[string]*model.APIPermission, len(defaultAPIBindings))
	for _, item := range defaultAPIBindings {
		defaults[apiRouteKey(item.Method, item.Path)] = item
	}

	itemsByRoute := make(map[string]*model.APIPermission)
	for _, route := range routes {
		path := normalizeRoutePath(route.Route)
		if !isPermissionControlledRoute(path) {
			continue
		}
		method := strings.ToUpper(route.Method)
		key := apiRouteKey(method, path)
		if _, exists := itemsByRoute[key]; exists {
			continue
		}
		item := &model.APIPermission{
			Method:  method,
			Path:    path,
			Service: apiService(path),
		}
		if configured, ok := defaults[key]; ok {
			item.Service = configured.Service
			item.PermissionCode = configured.PermissionCode
		}
		itemsByRoute[key] = item
	}

	items := make([]*model.APIPermission, 0, len(itemsByRoute))
	for _, item := range itemsByRoute {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Service != items[j].Service {
			return items[i].Service < items[j].Service
		}
		if items[i].Path != items[j].Path {
			return items[i].Path < items[j].Path
		}
		return items[i].Method < items[j].Method
	})
	return items
}

// EnsureAPIRegistry 幂等登记路由表中的全部受权限保护 API。
func EnsureAPIRegistry(ctx context.Context, store *repository.Store, routes []ghttp.RouterItem) ([]*model.APIPermission, error) {
	items := DiscoverRegisteredAPIs(routes)
	for _, item := range items {
		if err := store.EnsureAPIPermission(ctx, item); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func normalizeRoutePath(path string) string {
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

func isPermissionControlledRoute(path string) bool {
	if !strings.HasPrefix(path, "/api/") {
		return false
	}
	switch path {
	case "/api/health", "/api/me", "/api/access/check", "/api/access/user-permission":
		return false
	}
	return !strings.HasPrefix(path, "/api/auth/") && !strings.HasPrefix(path, "/api/public/")
}

func apiService(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) >= 2 && segments[0] == "api" {
		return segments[1]
	}
	return "admin"
}

func apiRouteKey(method string, path string) string {
	return strings.ToUpper(method) + " " + path
}

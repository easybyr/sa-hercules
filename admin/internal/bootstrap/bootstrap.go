package bootstrap

import (
	"context"
	"sort"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/example/sa-hercules/admin/pkg/password"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PermissionSeed 表示内置权限。
type PermissionSeed struct {
	Code string
	Name string
}

// SuperAdminConfig 是首次初始化超级管理员所需的凭据。
type SuperAdminConfig struct {
	Username string
	Password string
}

// BuiltinPermissions 是系统管理接口使用的权限编码。
var BuiltinPermissions = []PermissionSeed{
	{Code: consts.PermissionDashboardRead, Name: "查看仪表盘"},
	{Code: consts.PermissionTenantRead, Name: "查看租户"},
	{Code: consts.PermissionTenantCreate, Name: "创建租户"},
	{Code: consts.PermissionTenantUpdate, Name: "修改租户"},
	{Code: consts.PermissionTenantDelete, Name: "删除租户"},
	{Code: consts.PermissionUserRead, Name: "查看用户"},
	{Code: consts.PermissionUserCreate, Name: "创建用户"},
	{Code: consts.PermissionUserUpdate, Name: "修改用户"},
	{Code: consts.PermissionUserDelete, Name: "删除用户"},
	{Code: consts.PermissionUserAssignRole, Name: "分配用户角色"},
	{Code: consts.PermissionUserAssignPermission, Name: "分配用户权限"},
	{Code: consts.PermissionUserAssignTenant, Name: "分配用户租户"},
	{Code: consts.PermissionPermissionRead, Name: "查看权限"},
	{Code: consts.PermissionPermissionCreate, Name: "创建权限"},
	{Code: consts.PermissionPermissionUpdate, Name: "修改权限"},
	{Code: consts.PermissionPermissionDelete, Name: "删除权限"},
	{Code: consts.PermissionRoleRead, Name: "查看角色"},
	{Code: consts.PermissionRoleCreate, Name: "创建角色"},
	{Code: consts.PermissionRoleUpdate, Name: "修改角色"},
	{Code: consts.PermissionRoleDelete, Name: "删除角色"},
	{Code: consts.PermissionRoleAssignPermission, Name: "分配角色权限"},
}

// EnsureSuperAdmin 幂等初始化超级管理员、根租户、角色和权限。
func EnsureSuperAdmin(ctx context.Context, store *repository.Store, routes []ghttp.RouterItem, adminConfig SuperAdminConfig) error {
	registeredAPIs, err := EnsureAPIRegistry(ctx, store, routes)
	if err != nil {
		return err
	}
	tenant, err := store.FindTenantByCode(ctx, consts.SuperAdminTenantCode)
	if err != nil {
		return err
	}
	if tenant.ID == 0 {
		tenant = &model.Tenant{
			Code: consts.SuperAdminTenantCode, Name: "系统管理租户", Description: "超级管理员根租户",
			Status: int(consts.StatusNormal), Version: 1,
		}
		tenant.ID, err = store.CreateTenant(ctx, tenant)
		if err != nil {
			return err
		}
	}

	user, err := store.FindUserByUID(ctx, consts.SuperAdminUID)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		hashed, hashErr := password.Hash(adminConfig.Password)
		if hashErr != nil {
			return hashErr
		}
		user = &model.User{
			UID: consts.SuperAdminUID, Username: adminConfig.Username, Nickname: "超级管理员",
			PasswordHash: hashed,
			Status:       int(consts.StatusNormal), Version: 1,
		}
		if _, err = store.CreateUser(ctx, user, tenant.ID); err != nil {
			return err
		}
	} else if err = store.AddUserTenant(ctx, consts.SuperAdminUID, tenant.ID); err != nil {
		return err
	}

	role, err := store.FindRoleByCode(ctx, tenant.ID, consts.SuperAdminRoleCode)
	if err != nil {
		return err
	}
	if role.ID == 0 {
		role = &model.Role{
			Code: consts.SuperAdminRoleCode, Name: "超级管理员", Description: "系统内置超级管理员角色",
			TenantID: tenant.ID, Status: 1, IsSystem: 1, Version: 1,
		}
		role.ID, err = store.CreateRole(ctx, role)
		role.RoleID = role.ID
		if err != nil {
			return err
		}
	}
	if err = store.EnsureUserRole(ctx, consts.SuperAdminUID, role.ID); err != nil {
		return err
	}

	codes := make([]string, 0, len(BuiltinPermissions))
	for _, seed := range BuiltinPermissions {
		nodeType := consts.PermissionNodeTypeForCode(seed.Code)
		apis := registeredPathsForPermission(seed.Code, registeredAPIs)
		permission, findErr := store.FindPermission(ctx, tenant.ID, seed.Code)
		if findErr != nil {
			return findErr
		}
		if permission.ID == 0 {
			permission = &model.Permission{
				Code: seed.Code, NodeType: nodeType, APIs: apis,
				Name: seed.Name, Description: "系统内置权限",
				TenantID: tenant.ID, Status: 1, Version: 1,
			}
			permission.ID, err = store.CreatePermission(ctx, permission)
			if err != nil {
				return err
			}
		}
		if err = store.SavePermissionNode(ctx, "", seed.Code, nodeType, apis); err != nil {
			return err
		}
		codes = append(codes, seed.Code)
	}
	if err = EnsureTenantAdminRoles(ctx, store); err != nil {
		return err
	}
	return store.ReplaceRolePermissions(ctx, role.ID, codes)
}

// EnsureTenantAdminRoles 为已有租户补齐内置租户管理员角色及其全部权限。
func EnsureTenantAdminRoles(ctx context.Context, store *repository.Store) error {
	tenants, err := store.ListTenants(ctx)
	if err != nil {
		return err
	}
	for _, tenant := range tenants {
		if _, err = store.EnsureTenantAdminRole(ctx, tenant.ID); err != nil {
			return err
		}
	}
	return nil
}

func registeredPathsForPermission(code string, registeredAPIs []*model.APIPermission) []string {
	paths := make([]string, 0)
	for _, item := range registeredAPIs {
		if item.PermissionCode == code {
			paths = append(paths, item.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

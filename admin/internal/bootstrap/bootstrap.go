package bootstrap

import (
	"context"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/example/sa-hercules/admin/pkg/password"
)

// PermissionSeed 表示内置权限。
type PermissionSeed struct {
	Code string
	Name string
}

// BuiltinPermissions 是系统管理接口使用的权限编码。
var BuiltinPermissions = []PermissionSeed{
	{Code: "dashboard.read", Name: "查看仪表盘"},
	{Code: "organization.read", Name: "查看组织"},
	{Code: "organization.create", Name: "创建组织"},
	{Code: "organization.update", Name: "修改组织"},
	{Code: "organization.delete", Name: "删除组织"},
	{Code: "user.read", Name: "查看用户"},
	{Code: "user.create", Name: "创建用户"},
	{Code: "user.update", Name: "修改用户"},
	{Code: "user.delete", Name: "删除用户"},
	{Code: "user.assign_role", Name: "分配用户角色"},
	{Code: "user.assign_permission", Name: "分配用户权限"},
	{Code: "user.assign_organization", Name: "分配用户组织"},
	{Code: "permission.read", Name: "查看权限"},
	{Code: "permission.create", Name: "创建权限"},
	{Code: "permission.update", Name: "修改权限"},
	{Code: "permission.delete", Name: "删除权限"},
	{Code: "role.read", Name: "查看角色"},
	{Code: "role.create", Name: "创建角色"},
	{Code: "role.update", Name: "修改角色"},
	{Code: "role.delete", Name: "删除角色"},
	{Code: "role.assign_permission", Name: "分配角色权限"},
}

// EnsureSuperAdmin 幂等初始化超级管理员、根组织、角色和权限。
func EnsureSuperAdmin(ctx context.Context, store *repository.Store) error {
	organization, err := store.FindOrganizationByOrgID(ctx, consts.SuperAdminOrgID)
	if err != nil {
		return err
	}
	if organization.ID == 0 {
		organization = &model.Organization{
			OrgID: consts.SuperAdminOrgID, Name: "系统管理组织", Description: "超级管理员根组织",
			Status: int(consts.StatusNormal), Version: 1,
		}
		if _, err = store.CreateOrganization(ctx, organization); err != nil {
			return err
		}
	}

	user, err := store.FindUserByUID(ctx, consts.SuperAdminUID)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		hashed, hashErr := password.Hash("admin")
		if hashErr != nil {
			return hashErr
		}
		user = &model.User{
			UID: consts.SuperAdminUID, Username: "admin", Nickname: "超级管理员",
			Password: hashed, OrgID: consts.SuperAdminOrgID,
			Status: int(consts.StatusNormal), Version: 1,
		}
		if _, err = store.CreateUser(ctx, user); err != nil {
			return err
		}
	}

	role, err := store.FindRole(ctx, consts.SuperAdminOrgID, consts.SuperAdminRoleID)
	if err != nil {
		return err
	}
	if role.ID == 0 {
		role = &model.Role{
			RoleID: consts.SuperAdminRoleID, Name: "超级管理员",
			Description: "系统内置超级管理员角色", OrgID: consts.SuperAdminOrgID, Version: 1,
		}
		if _, err = store.CreateRole(ctx, role); err != nil {
			return err
		}
	}
	if err = store.EnsureUserRole(ctx, consts.SuperAdminUID, consts.SuperAdminRoleID); err != nil {
		return err
	}

	codes := make([]string, 0, len(BuiltinPermissions))
	for _, seed := range BuiltinPermissions {
		permission, findErr := store.FindPermission(ctx, consts.SuperAdminOrgID, seed.Code)
		if findErr != nil {
			return findErr
		}
		if permission.ID == 0 {
			permission = &model.Permission{
				Code: seed.Code, Name: seed.Name, Description: "系统内置权限",
				OrgID: consts.SuperAdminOrgID, Version: 1,
			}
			if _, err = store.CreatePermission(ctx, permission); err != nil {
				return err
			}
		}
		codes = append(codes, seed.Code)
	}
	return store.ReplaceRolePermissions(ctx, consts.SuperAdminRoleID, codes)
}

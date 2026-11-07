package repository

import (
	"context"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *Store) UserRoles(ctx context.Context, uid, tenantID int64) ([]*model.Role, error) {
	items := make([]*model.Role, 0)
	err := g.DB().Model("sys_role r").Ctx(ctx).Fields("r.*").
		InnerJoin("sys_user_role ur", "ur.tenant_id = r.tenant_id AND ur.role_id = r.id").
		InnerJoin("sys_user u", "u.id = ur.user_id").
		InnerJoin("sys_tenant_member tm", "tm.id = ur.tenant_member_id AND tm.tenant_id = ur.tenant_id AND tm.user_id = ur.user_id").
		Where("u.uid", uid).Where("r.tenant_id", tenantID).Where("tm.status", 1).Where("r.status", 1).
		Where("r.code <> ?", directRoleCode(uid)).
		OrderAsc("r.name").Scan(&items)
	for _, item := range items {
		item.RoleID = item.ID
	}
	return items, err
}

func (s *Store) UserPermissions(ctx context.Context, uid, tenantID int64) ([]*model.Permission, error) {
	items := make([]*model.Permission, 0)
	err := g.DB().Model("sys_permission p").Ctx(ctx).Fields("p.*").
		InnerJoin("sys_role_permission rp", "rp.tenant_id = p.tenant_id AND rp.permission_id = p.id").
		InnerJoin("sys_role r", "r.id = rp.role_id AND r.tenant_id = rp.tenant_id").
		InnerJoin("sys_user_role ur", "ur.role_id = r.id AND ur.tenant_id = r.tenant_id").
		InnerJoin("sys_user u", "u.id = ur.user_id").
		InnerJoin("sys_tenant_member tm", "tm.id = ur.tenant_member_id AND tm.tenant_id = ur.tenant_id AND tm.user_id = ur.user_id").
		Where("u.uid", uid).Where("p.tenant_id", tenantID).Where("p.status", 1).Where("r.status", 1).Where("tm.status", 1).
		OrderAsc("p.code").Scan(&items)
	if err != nil {
		return nil, err
	}
	unique := make([]*model.Permission, 0, len(items))
	seen := make(map[int64]bool, len(items))
	for _, item := range items {
		if !seen[item.ID] {
			seen[item.ID] = true
			unique = append(unique, item)
		}
	}
	return unique, nil
}

func (s *Store) UserHasPermission(ctx context.Context, tenantID, uid int64, code string) (bool, error) {
	count, err := g.DB().Model("sys_permission p").Ctx(ctx).
		InnerJoin("sys_role_permission rp", "rp.tenant_id = p.tenant_id AND rp.permission_id = p.id").
		InnerJoin("sys_role r", "r.id = rp.role_id AND r.tenant_id = rp.tenant_id").
		InnerJoin("sys_user_role ur", "ur.role_id = r.id AND ur.tenant_id = r.tenant_id").
		InnerJoin("sys_user u", "u.id = ur.user_id").
		InnerJoin("sys_tenant_member tm", "tm.id = ur.tenant_member_id AND tm.tenant_id = ur.tenant_id AND tm.user_id = ur.user_id").
		Where("u.uid", uid).Where("p.tenant_id", tenantID).Where("p.code", code).
		Where("p.status", 1).Where("r.status", 1).Where("tm.status", 1).Count()
	return count > 0, err
}

func (s *Store) ReplaceUserRoles(ctx context.Context, uid, tenantID int64, roleIDs []int64) error {
	return s.ReplaceUserRolesWithLog(ctx, uid, tenantID, roleIDs, nil)
}

func (s *Store) ReplaceUserRolesWithLog(ctx context.Context, uid, tenantID int64, roleIDs []int64, operation *model.PermissionOperationLog) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user model.User
		if err := tx.Model("sys_user").Where("uid", uid).Scan(&user); err != nil {
			return err
		}
		var member model.TenantMember
		if err := tx.Model("sys_tenant_member").Where("tenant_id", tenantID).Where("user_id", user.ID).Scan(&member); err != nil {
			return err
		}
		personalRole := &model.Role{}
		if err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("code", directRoleCode(uid)).Scan(personalRole); err != nil {
			return err
		}
		deleteRoles := tx.Model("sys_user_role").Where("tenant_id", tenantID).Where("tenant_member_id", member.ID)
		if personalRole.ID != 0 {
			deleteRoles = deleteRoles.Where("role_id <> ?", personalRole.ID)
		}
		if _, err := deleteRoles.Delete(); err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			var role model.Role
			if err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("id", roleID).Where("status", 1).Scan(&role); err != nil {
				return err
			}
			if role.ID == 0 {
				continue
			}
			row := model.UserRole{TenantID: tenantID, TenantMemberID: member.ID, UserID: user.ID, RoleID: roleID}
			if _, err := tx.Model("sys_user_role").Data(row).Insert(); err != nil {
				return err
			}
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
}

func (s *Store) EnsureUserRole(ctx context.Context, uid, roleID int64) error {
	var role model.Role
	if err := g.DB().Model("sys_role").Ctx(ctx).Where("id", roleID).Scan(&role); err != nil {
		return err
	}
	if role.ID == 0 {
		return nil
	}
	var user model.User
	if err := g.DB().Model("sys_user").Ctx(ctx).Where("uid", uid).Scan(&user); err != nil {
		return err
	}
	var member model.TenantMember
	if err := g.DB().Model("sys_tenant_member").Ctx(ctx).Where("tenant_id", role.TenantID).Where("user_id", user.ID).Scan(&member); err != nil {
		return err
	}
	count, err := g.DB().Model("sys_user_role").Ctx(ctx).Where("tenant_id", role.TenantID).Where("tenant_member_id", member.ID).Where("role_id", roleID).Count()
	if err != nil || count > 0 {
		return err
	}
	_, err = g.DB().Model("sys_user_role").Ctx(ctx).Data(model.UserRole{TenantID: role.TenantID, TenantMemberID: member.ID, UserID: user.ID, RoleID: roleID}).Insert()
	return err
}

func (s *Store) ReplaceRolePermissions(ctx context.Context, roleID int64, codes []string) error {
	return s.ReplaceRolePermissionsWithLog(ctx, roleID, codes, nil)
}

func (s *Store) ReplaceRolePermissionsWithLog(ctx context.Context, roleID int64, codes []string, operation *model.PermissionOperationLog) error {
	var role model.Role
	if err := g.DB().Model("sys_role").Ctx(ctx).Where("id", roleID).Scan(&role); err != nil {
		return err
	}
	if role.ID == 0 {
		return nil
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("sys_role_permission").Where("tenant_id", role.TenantID).Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		if err := insertRolePermissions(tx, role.TenantID, roleID, codes); err != nil {
			return err
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
}

func (s *Store) RolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	values, err := g.DB().Model("sys_role_permission rp").Ctx(ctx).
		InnerJoin("sys_permission p", "p.id = rp.permission_id AND p.tenant_id = rp.tenant_id").
		Where("rp.role_id", roleID).OrderAsc("p.code").Array("p.code")
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(values))
	for _, value := range values {
		codes = append(codes, value.String())
	}
	return codes, nil
}

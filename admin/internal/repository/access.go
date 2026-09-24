package repository

import (
	"context"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/idgen"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserRoles 查询用户在其组织内的角色。
func (s *Store) UserRoles(ctx context.Context, uid int64, orgID int64) ([]*model.Role, error) {
	items := make([]*model.Role, 0)
	err := g.DB().Model("role r").Ctx(ctx).
		InnerJoin("user_role ur", "ur.role_id = r.role_id").
		Where("ur.uid", uid).Where("r.org_id", orgID).WhereGT("r.role_id", 0).
		Fields("r.*").OrderAsc("r.role_id").Scan(&items)
	return items, err
}

// UserPermissions 查询用户在其当前组织内通过角色获得的权限。
func (s *Store) UserPermissions(ctx context.Context, uid int64, orgID int64) ([]*model.Permission, error) {
	items := make([]*model.Permission, 0)
	err := g.DB().Model("permission p").Ctx(ctx).
		InnerJoin("role_permission rp", "rp.permission_code = p.code").
		InnerJoin("role r", "r.role_id = rp.role_id AND r.org_id = p.org_id").
		InnerJoin("user_role ur", "ur.role_id = r.role_id").
		Where("ur.uid", uid).Where("p.org_id", orgID).
		Fields("DISTINCT p.*").OrderAsc("p.code").Scan(&items)
	return items, err
}

// ReplaceUserRoles 原子替换用户角色，并清空旧的用户直接权限。
func (s *Store) ReplaceUserRoles(ctx context.Context, uid int64, roleIDs []int64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("user_role").Where("uid", uid).WhereGT("role_id", 0).Delete(); err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			item := &model.UserRole{UID: uid, RoleID: roleID, Version: 1}
			if _, err := tx.Model("user_role").Data(item).Insert(); err != nil {
				return err
			}
		}
		if _, err := tx.Model("role_permission").Where("role_id", idgen.PersonalRoleID(uid)).Delete(); err != nil {
			return err
		}
		return nil
	})
}

// EnsureUserRole 确保用户角色关联存在。
func (s *Store) EnsureUserRole(ctx context.Context, uid int64, roleID int64) error {
	count, err := g.DB().Model("user_role").Ctx(ctx).
		Where("uid", uid).Where("role_id", roleID).Count()
	if err != nil || count > 0 {
		return err
	}
	item := &model.UserRole{UID: uid, RoleID: roleID, Version: 1}
	_, err = g.DB().Model("user_role").Ctx(ctx).Data(item).Insert()
	return err
}

// ReplaceRolePermissions 原子替换角色权限。
func (s *Store) ReplaceRolePermissions(ctx context.Context, roleID int64, codes []string) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("role_permission").Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		for _, code := range codes {
			item := &model.RolePermission{
				RoleID: roleID, PermissionCode: code, Version: 1,
			}
			if _, err := tx.Model("role_permission").Data(item).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// RolePermissions 查询角色权限编码。
func (s *Store) RolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	values, err := g.DB().Model("role_permission").Ctx(ctx).
		Where("role_id", roleID).OrderAsc("permission_code").Array("permission_code")
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(values))
	for _, value := range values {
		codes = append(codes, value.String())
	}
	return codes, nil
}

// SetUserOrganization 修改用户组织并清理旧角色。
func (s *Store) SetUserOrganization(ctx context.Context, uid int64, orgID int64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("user_role").Where("uid", uid).Delete(); err != nil {
			return err
		}
		personalRoleID := idgen.PersonalRoleID(uid)
		if _, err := tx.Model("role_permission").Where("role_id", personalRoleID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("role").Where("role_id", personalRoleID).Delete(); err != nil {
			return err
		}
		type update struct {
			OrgID   int64   `db:"org_id"`
			Version gdb.Raw `db:"version"`
		}
		_, err := tx.Model("user").Where("uid", uid).
			Data(update{OrgID: orgID, Version: gdb.Raw("version + 1")}).Update()
		return err
	})
}

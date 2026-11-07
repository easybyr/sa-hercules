package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *Store) ListRoles(ctx context.Context, tenantID int64, includeSystem bool, keyword string, page model.PageRequest) ([]*model.Role, int, error) {
	items := make([]*model.Role, 0)
	db := g.DB().Model("sys_role").Ctx(ctx)
	if tenantID != 0 {
		db = db.Where("tenant_id", tenantID)
	}
	if !includeSystem {
		db = db.Where("is_system", 0)
	}
	db = db.Where("code NOT REGEXP ?", "^user_[0-9]+_direct$")
	if keyword != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	for _, item := range items {
		item.RoleID = item.ID
	}
	return items, total, err
}

// EnsureTenantAdminRole 创建或修复租户管理员角色，并令其拥有该租户全部权限。
func (s *Store) EnsureTenantAdminRole(ctx context.Context, tenantID int64) (*model.Role, error) {
	var role *model.Role
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var err error
		role, err = ensureTenantAdminRoleTx(tx, tenantID)
		return err
	})
	return role, err
}

func ensureTenantAdminRoleTx(tx gdb.TX, tenantID int64) (*model.Role, error) {
	var role model.Role
	err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("code", consts.TenantAdminRoleCode).Scan(&role)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if role.ID == 0 {
		role = model.Role{
			TenantID: tenantID, Code: consts.TenantAdminRoleCode, Name: "租户管理员",
			Description: "系统内置租户管理员，拥有该租户的全部权限", Status: 1, IsSystem: 1, Version: 1,
		}
		id, err := createRoleTx(tx, &role)
		if err != nil {
			return nil, err
		}
		role.ID, role.RoleID = id, id
	} else if role.IsSystem == 0 {
		if _, err := tx.Model("sys_role").Where("id", role.ID).Data(map[string]any{"is_system": 1}).Update(); err != nil {
			return nil, err
		}
		role.IsSystem = 1
	}
	if err := syncTenantAdminPermissionsTx(tx, tenantID, role.ID); err != nil {
		return nil, err
	}
	role.RoleID = role.ID
	return &role, nil
}

func syncTenantAdminPermissionsTx(tx gdb.TX, tenantID, roleID int64) error {
	if _, err := tx.Model("sys_role_permission").Where("tenant_id", tenantID).Where("role_id", roleID).Delete(); err != nil {
		return err
	}
	permissionIDs, err := tx.Model("sys_permission").Where("tenant_id", tenantID).Array("id")
	if err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if _, err := tx.Model("sys_role_permission").Data(model.RolePermission{
			TenantID: tenantID, RoleID: roleID, PermissionID: permissionID.Int64(),
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) FindRole(ctx context.Context, tenantID int64, roleID int64) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("sys_role").Ctx(ctx).Where("tenant_id", tenantID).Where("id", roleID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	item.RoleID = item.ID
	return &item, err
}

func (s *Store) FindRoleByRoleID(ctx context.Context, roleID int64) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("sys_role").Ctx(ctx).Where("id", roleID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	item.RoleID = item.ID
	return &item, err
}

func (s *Store) FindRoleByCode(ctx context.Context, tenantID int64, code string) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("sys_role").Ctx(ctx).Where("tenant_id", tenantID).Where("code", code).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	item.RoleID = item.ID
	return &item, err
}

func (s *Store) FindRoleByName(ctx context.Context, tenantID int64, name string) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("sys_role").Ctx(ctx).Where("tenant_id", tenantID).Where("name", name).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	item.RoleID = item.ID
	return &item, err
}

func createRoleTx(tx gdb.TX, item *model.Role) (int64, error) {
	record := struct {
		TenantID    int64  `db:"tenant_id"`
		Code        string `db:"code"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Status      int    `db:"status"`
		IsSystem    int    `db:"is_system"`
		Version     int    `db:"version"`
	}{item.TenantID, item.Code, item.Name, item.Description, item.Status, item.IsSystem, item.Version}
	result, err := tx.Model("sys_role").Data(record).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) CreateRole(ctx context.Context, item *model.Role) (int64, error) {
	var id int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var err error
		id, err = createRoleTx(tx, item)
		return err
	})
	return id, err
}

func (s *Store) CreateRoleWithPermissions(ctx context.Context, item *model.Role, permissionCodes []string) (int64, error) {
	return s.CreateRoleWithPermissionsAndLog(ctx, item, permissionCodes, nil)
}

func (s *Store) CreateRoleWithPermissionsAndLog(ctx context.Context, item *model.Role, permissionCodes []string, operation *model.PermissionOperationLog) (int64, error) {
	var id int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var err error
		id, err = createRoleTx(tx, item)
		if err != nil {
			return err
		}
		item.ID, item.RoleID = id, id
		if err = insertRolePermissions(tx, item.TenantID, id, permissionCodes); err != nil {
			return err
		}
		if operation != nil {
			operation.TargetID = id
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
	return id, err
}

func (s *Store) UpdateRole(ctx context.Context, tenantID int64, roleID int64, name string, description string, version int) (int64, error) {
	record := struct {
		Name        string `db:"name"`
		Description string `db:"description"`
		Version     int    `db:"version"`
	}{name, description, version + 1}
	result, err := g.DB().Model("sys_role").Ctx(ctx).Where("tenant_id", tenantID).Where("id", roleID).Where("version", version).Data(record).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) UpdateRoleWithPermissions(ctx context.Context, tenantID int64, roleID int64, name, description string, version int, codes []string) (int64, error) {
	return s.UpdateRoleWithPermissionsAndLog(ctx, tenantID, roleID, name, description, version, codes, nil)
}

func (s *Store) UpdateRoleWithPermissionsAndLog(ctx context.Context, tenantID int64, roleID int64, name, description string, version int, codes []string, operation *model.PermissionOperationLog) (int64, error) {
	var affected int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record := struct {
			Name        string `db:"name"`
			Description string `db:"description"`
			Version     int    `db:"version"`
		}{name, description, version + 1}
		result, err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("id", roleID).Where("version", version).Data(record).Update()
		if err != nil {
			return err
		}
		affected, err = result.RowsAffected()
		if err != nil || affected == 0 {
			return err
		}
		if _, err = tx.Model("sys_role_permission").Where("tenant_id", tenantID).Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		if err = insertRolePermissions(tx, tenantID, roleID, codes); err != nil {
			return err
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
	return affected, err
}

func insertRolePermissions(tx gdb.TX, tenantID, roleID int64, permissionCodes []string) error {
	for _, code := range permissionCodes {
		var permission model.Permission
		if err := tx.Model("sys_permission").Where("tenant_id", tenantID).Where("code", code).Scan(&permission); err != nil {
			return err
		}
		if permission.ID == 0 {
			continue
		}
		record := model.RolePermission{TenantID: tenantID, RoleID: roleID, PermissionID: permission.ID}
		if _, err := tx.Model("sys_role_permission").Data(record).Insert(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeleteRole(ctx context.Context, tenantID int64, roleID int64) error {
	return s.DeleteRoleWithLog(ctx, tenantID, roleID, nil)
}

func (s *Store) DeleteRoleWithLog(ctx context.Context, tenantID int64, roleID int64, operation *model.PermissionOperationLog) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("sys_user_role").Where("tenant_id", tenantID).Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_role_permission").Where("tenant_id", tenantID).Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("id", roleID).Delete(); err != nil {
			return err
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
}

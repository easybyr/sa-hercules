package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ListRoles 查询角色，orgID为0时查询全部。
func (s *Store) ListRoles(
	ctx context.Context,
	orgID int64,
	includeSystem bool,
	keyword string,
	page model.PageRequest,
) ([]*model.Role, int, error) {
	items := make([]*model.Role, 0)
	db := g.DB().Model("role").Ctx(ctx)
	if orgID != 0 {
		db = db.Where("org_id", orgID)
	}
	if !includeSystem {
		db = db.WhereGT("role_id", 0)
	}
	if keyword != "" {
		db = db.WhereLike("name", "%"+keyword+"%")
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

// FindRole 查询角色。
func (s *Store) FindRole(ctx context.Context, orgID int64, roleID int64) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("role").Ctx(ctx).
		Where("org_id", orgID).Where("role_id", roleID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// FindRoleByRoleID 按全局角色业务ID查询角色。
func (s *Store) FindRoleByRoleID(ctx context.Context, roleID int64) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("role").Ctx(ctx).Where("role_id", roleID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// FindRoleByName 查询组织内同名角色。
func (s *Store) FindRoleByName(ctx context.Context, orgID int64, name string) (*model.Role, error) {
	var item model.Role
	err := g.DB().Model("role").Ctx(ctx).
		Where("org_id", orgID).Where("name", name).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// CreateRole 新建角色。
func (s *Store) CreateRole(ctx context.Context, item *model.Role) (int64, error) {
	result, err := g.DB().Model("role").Ctx(ctx).Data(item).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CreateRoleWithPermissions 原子创建角色及其权限关联。
func (s *Store) CreateRoleWithPermissions(
	ctx context.Context,
	item *model.Role,
	permissionCodes []string,
) (int64, error) {
	var id int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model("role").Data(item).Insert()
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		return insertRolePermissions(tx, item.RoleID, permissionCodes)
	})
	return id, err
}

// UpdateRole 更新角色。
func (s *Store) UpdateRole(
	ctx context.Context,
	orgID int64,
	roleID int64,
	name string,
	description string,
	version int,
) (int64, error) {
	type update struct {
		Name        string `db:"name"`
		Description string `db:"description"`
		Version     int    `db:"version"`
	}
	result, err := g.DB().Model("role").Ctx(ctx).
		Where("org_id", orgID).Where("role_id", roleID).Where("version", version).
		Data(update{Name: name, Description: description, Version: version + 1}).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// UpdateRoleWithPermissions 使用乐观锁原子更新角色及权限关联。
func (s *Store) UpdateRoleWithPermissions(
	ctx context.Context,
	orgID int64,
	roleID int64,
	name string,
	description string,
	version int,
	permissionCodes []string,
) (int64, error) {
	type update struct {
		Name        string `db:"name"`
		Description string `db:"description"`
		Version     int    `db:"version"`
	}
	var affected int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model("role").
			Where("org_id", orgID).Where("role_id", roleID).Where("version", version).
			Data(update{Name: name, Description: description, Version: version + 1}).Update()
		if err != nil {
			return err
		}
		affected, err = result.RowsAffected()
		if err != nil || affected == 0 {
			return err
		}
		if _, err = tx.Model("role_permission").Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		return insertRolePermissions(tx, roleID, permissionCodes)
	})
	return affected, err
}

func insertRolePermissions(tx gdb.TX, roleID int64, permissionCodes []string) error {
	for _, code := range permissionCodes {
		item := &model.RolePermission{
			RoleID: roleID, PermissionCode: code, Version: 1,
		}
		if _, err := tx.Model("role_permission").Data(item).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRole 删除角色及关联。
func (s *Store) DeleteRole(ctx context.Context, orgID int64, roleID int64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("user_role").Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("role_permission").Where("role_id", roleID).Delete(); err != nil {
			return err
		}
		_, err := tx.Model("role").Where("org_id", orgID).Where("role_id", roleID).Delete()
		return err
	})
}

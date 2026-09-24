package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ListPermissions 分页查询权限，orgID为0时查询全部。
func (s *Store) ListPermissions(
	ctx context.Context,
	orgID int64,
	keyword string,
	page model.PageRequest,
) ([]*model.Permission, int, error) {
	items := make([]*model.Permission, 0)
	db := g.DB().Model("permission").Ctx(ctx)
	if orgID != 0 {
		db = db.Where("org_id", orgID)
	}
	if keyword != "" {
		db = db.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

// FindPermission 查询组织内权限。
func (s *Store) FindPermission(ctx context.Context, orgID int64, code string) (*model.Permission, error) {
	var item model.Permission
	err := g.DB().Model("permission").Ctx(ctx).
		Where("org_id", orgID).Where("code", code).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// FindPermissionByName 查询组织内同名权限。
func (s *Store) FindPermissionByName(ctx context.Context, orgID int64, name string) (*model.Permission, error) {
	var item model.Permission
	err := g.DB().Model("permission").Ctx(ctx).
		Where("org_id", orgID).Where("name", name).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// CreatePermission 新建权限。
func (s *Store) CreatePermission(ctx context.Context, item *model.Permission) (int64, error) {
	result, err := g.DB().Model("permission").Ctx(ctx).Data(item).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdatePermission 更新权限。
func (s *Store) UpdatePermission(
	ctx context.Context,
	orgID int64,
	code string,
	name string,
	description string,
	version int,
) (int64, error) {
	type update struct {
		Name        string `db:"name"`
		Description string `db:"description"`
		Version     int    `db:"version"`
	}
	result, err := g.DB().Model("permission").Ctx(ctx).
		Where("org_id", orgID).Where("code", code).Where("version", version).
		Data(update{Name: name, Description: description, Version: version + 1}).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeletePermission 删除权限及角色关联。
func (s *Store) DeletePermission(ctx context.Context, orgID int64, code string) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		roleIDs, err := tx.Model("role").Where("org_id", orgID).Array("role_id")
		if err != nil {
			return err
		}
		if len(roleIDs) > 0 {
			if _, err = tx.Model("role_permission").
				WhereIn("role_id", roleIDs).Where("permission_code", code).Delete(); err != nil {
				return err
			}
		}
		_, err = tx.Model("permission").Where("org_id", orgID).Where("code", code).Delete()
		return err
	})
}

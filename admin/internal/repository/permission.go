package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *Store) ListPermissions(ctx context.Context, tenantID int64, keyword string, page model.PageRequest) ([]*model.Permission, int, error) {
	items := make([]*model.Permission, 0)
	db := g.DB().Model("sys_permission").Ctx(ctx)
	if tenantID != 0 {
		db = db.Where("tenant_id", tenantID)
	}
	if keyword != "" {
		term := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", term, term, term)
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderAsc("parent_id").OrderAsc("node_type").OrderAsc("name").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

func (s *Store) FindPermission(ctx context.Context, tenantID int64, code string) (*model.Permission, error) {
	var item model.Permission
	err := g.DB().Model("sys_permission").Ctx(ctx).Where("tenant_id", tenantID).Where("code", code).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) FindPermissionByName(ctx context.Context, tenantID int64, name string) (*model.Permission, error) {
	var item model.Permission
	err := g.DB().Model("sys_permission").Ctx(ctx).Where("tenant_id", tenantID).Where("name", name).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) CountPermissionsByCode(ctx context.Context, code string) (int, error) {
	return g.DB().Model("sys_permission").Ctx(ctx).Where("code", code).Count()
}

func permissionRecord(item *model.Permission) map[string]any {
	return map[string]any{
		"tenant_id": item.TenantID, "code": item.Code, "name": item.Name,
		"node_type": string(item.NodeType), "parent_id": item.ParentID,
		"description": item.Description, "status": 1, "version": item.Version,
	}
}

func (s *Store) CreatePermission(ctx context.Context, item *model.Permission) (int64, error) {
	var id int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model("sys_permission").Data(permissionRecord(item)).Insert()
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		item.ID = id
		_, err = ensureTenantAdminRoleTx(tx, item.TenantID)
		return err
	})
	return id, err
}

func (s *Store) CreatePermissionWithLog(ctx context.Context, item *model.Permission, operation *model.PermissionOperationLog, apis []string) (int64, error) {
	var id int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model("sys_permission").Data(permissionRecord(item)).Insert()
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		item.ID = id
		if _, err = ensureTenantAdminRoleTx(tx, item.TenantID); err != nil {
			return err
		}
		if err = replacePermissionNodeTx(tx, "", item.Code, item.NodeType, apis); err != nil {
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

func (s *Store) UpdatePermission(ctx context.Context, tenantID int64, oldCode string, item *model.Permission, version int) (int64, error) {
	return s.UpdatePermissionWithLog(ctx, tenantID, oldCode, item, version, nil)
}

func (s *Store) UpdatePermissionWithLog(ctx context.Context, tenantID int64, oldCode string, item *model.Permission, version int, operation *model.PermissionOperationLog) (int64, error) {
	var affected int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record := map[string]any{
			"code": item.Code, "name": item.Name, "node_type": string(item.NodeType),
			"parent_id": item.ParentID, "description": item.Description, "version": version + 1,
		}
		result, err := tx.Model("sys_permission").Where("tenant_id", tenantID).Where("code", oldCode).Where("version", version).Data(record).Update()
		if err != nil {
			return err
		}
		affected, err = result.RowsAffected()
		if err != nil || affected == 0 {
			return err
		}
		if err = replacePermissionAPIsTx(tx, oldCode, item.Code, item.APIs); err != nil {
			return err
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
	return affected, err
}

func (s *Store) DeletePermission(ctx context.Context, tenantID int64, code string) error {
	return s.DeletePermissionWithLog(ctx, tenantID, code, nil)
}

func (s *Store) DeletePermissionWithLog(ctx context.Context, tenantID int64, code string, operation *model.PermissionOperationLog) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var item model.Permission
		if err := tx.Model("sys_permission").Where("tenant_id", tenantID).Where("code", code).Scan(&item); err != nil {
			return err
		}
		if item.ID == 0 {
			return nil
		}
		if _, err := tx.Model("sys_role_permission").Where("tenant_id", tenantID).Where("permission_id", item.ID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_permission").Where("tenant_id", tenantID).Where("id", item.ID).Delete(); err != nil {
			return err
		}
		var remaining int
		remaining, err := tx.Model("sys_permission").Where("code", code).Count()
		if err != nil {
			return err
		}
		if remaining == 0 {
			if _, err := tx.Model("sys_api_permission").Where("permission_code", code).Data(map[string]any{"permission_code": ""}).Update(); err != nil {
				return err
			}
		}
		if operation != nil {
			return insertPermissionOperationLog(tx, operation)
		}
		return nil
	})
}

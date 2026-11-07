package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *Store) FindTenantByID(ctx context.Context, id int64) (*model.Tenant, error) {
	var item model.Tenant
	err := g.DB().Model("sys_tenant").Ctx(ctx).Where("id", id).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) FindTenantByCode(ctx context.Context, code string) (*model.Tenant, error) {
	var item model.Tenant
	err := g.DB().Model("sys_tenant").Ctx(ctx).Where("code", code).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) SearchTenantsExact(ctx context.Context, keyword string, limit int) ([]*model.Tenant, error) {
	items := make([]*model.Tenant, 0)
	err := g.DB().Model("sys_tenant").Ctx(ctx).Where("status", 1).
		Where("name = ? OR code = ?", keyword, keyword).OrderAsc("name").Limit(limit).Scan(&items)
	return items, err
}

func (s *Store) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	items := make([]*model.Tenant, 0)
	err := g.DB().Model("sys_tenant").Ctx(ctx).OrderAsc("id").Scan(&items)
	return items, err
}

func (s *Store) ListTenantsPage(ctx context.Context, allowedIDs []int64, restricted bool, keyword string, page model.PageRequest) ([]*model.Tenant, int, error) {
	items := make([]*model.Tenant, 0)
	if restricted && len(allowedIDs) == 0 {
		return items, 0, nil
	}
	db := g.DB().Model("sys_tenant").Ctx(ctx)
	if restricted {
		db = db.WhereIn("id", allowedIDs)
	}
	if keyword != "" {
		term := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", term, term, term)
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

func (s *Store) CreateTenant(ctx context.Context, item *model.Tenant) (int64, error) {
	var id int64
	record := struct {
		Code        string  `db:"code"`
		ParentCode  *string `db:"parent_code"`
		Name        string  `db:"name"`
		Status      int     `db:"status"`
		Description string  `db:"description"`
		Version     int     `db:"version"`
	}{item.Code, item.ParentCode, item.Name, item.Status, item.Description, item.Version}
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := tx.Model("sys_tenant").Data(record).Insert()
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		item.ID = id
		_, err = ensureTenantAdminRoleTx(tx, id)
		return err
	})
	return id, err
}

func (s *Store) UpdateTenant(ctx context.Context, item *model.Tenant, version int) (int64, error) {
	record := struct {
		Name        string  `db:"name"`
		ParentCode  *string `db:"parent_code"`
		Description string  `db:"description"`
		Status      int     `db:"status"`
		Version     int     `db:"version"`
	}{item.Name, item.ParentCode, item.Description, item.Status, version + 1}
	result, err := g.DB().Model("sys_tenant").Ctx(ctx).Where("id", item.ID).Where("version", version).Data(record).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) DeleteTenant(ctx context.Context, id int64) (int64, error) {
	result, err := g.DB().Model("sys_tenant").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteTenantCascade 清理租户范围内的成员关系、角色、权限和 API 绑定后删除租户。
func (s *Store) DeleteTenantCascade(ctx context.Context, tenantID int64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		permissions := make([]*model.Permission, 0)
		if err := tx.Model("sys_permission").Where("tenant_id", tenantID).Scan(&permissions); err != nil {
			return err
		}
		if _, err := tx.Model("sys_user_role").Where("tenant_id", tenantID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_role_permission").Where("tenant_id", tenantID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_role").Where("tenant_id", tenantID).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model("sys_permission").Where("tenant_id", tenantID).Delete(); err != nil {
			return err
		}
		for _, permission := range permissions {
			remaining, err := tx.Model("sys_permission").Where("code", permission.Code).Count()
			if err != nil {
				return err
			}
			if remaining == 0 {
				if _, err = tx.Model("sys_api_permission").Where("permission_code", permission.Code).Data(map[string]any{"permission_code": ""}).Update(); err != nil {
					return err
				}
			}
		}
		if _, err := tx.Model("sys_tenant_member").Where("tenant_id", tenantID).Delete(); err != nil {
			return err
		}
		_, err := tx.Model("sys_tenant").Where("id", tenantID).Delete()
		return err
	})
}

func (s *Store) CountTenantChildren(ctx context.Context, id int64) (int, error) {
	tenant, err := s.FindTenantByID(ctx, id)
	if err != nil || tenant.ID == 0 {
		return 0, err
	}
	return g.DB().Model("sys_tenant").Ctx(ctx).Where("parent_code", tenant.Code).Count()
}

func (s *Store) CountTenantUsers(ctx context.Context, tenantID int64) (int, error) {
	return g.DB().Model("sys_tenant_member").Ctx(ctx).Where("tenant_id", tenantID).Count()
}

func (s *Store) CountTenantPermissions(ctx context.Context, tenantID int64) (int, error) {
	return g.DB().Model("sys_permission").Ctx(ctx).Where("tenant_id", tenantID).Count()
}

func (s *Store) CountTenantRoles(ctx context.Context, tenantID int64) (int, error) {
	return g.DB().Model("sys_role").Ctx(ctx).Where("tenant_id", tenantID).Count()
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// FindOrganizationByID 按主键查询组织。
func (s *Store) FindOrganizationByID(ctx context.Context, id int64) (*model.Organization, error) {
	var item model.Organization
	err := g.DB().Model("organization").Ctx(ctx).Where("id", id).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// FindOrganizationByOrgID 按业务ID查询组织。
func (s *Store) FindOrganizationByOrgID(ctx context.Context, orgID int64) (*model.Organization, error) {
	var item model.Organization
	err := g.DB().Model("organization").Ctx(ctx).Where("org_id", orgID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// SearchOrganizationsExact 按完整名称或组织ID搜索有效组织。
func (s *Store) SearchOrganizationsExact(
	ctx context.Context,
	keyword string,
	limit int,
) ([]*model.Organization, error) {
	items := make([]*model.Organization, 0)
	db := g.DB().Model("organization").Ctx(ctx).Where("status", 1)
	if orgID, err := strconv.ParseInt(keyword, 10, 64); err == nil {
		db = db.Where("name = ? OR org_id = ?", keyword, orgID)
	} else {
		db = db.Where("name", keyword)
	}
	err := db.OrderAsc("name").Limit(limit).Scan(&items)
	return items, err
}

// ListOrganizations 查询全部组织。
func (s *Store) ListOrganizations(ctx context.Context) ([]*model.Organization, error) {
	items := make([]*model.Organization, 0)
	err := g.DB().Model("organization").Ctx(ctx).OrderAsc("id").Scan(&items)
	return items, err
}

// ListOrganizationsPage 分页查询组织，restricted为true时仅查询allowedIDs。
func (s *Store) ListOrganizationsPage(
	ctx context.Context,
	allowedIDs []int64,
	restricted bool,
	keyword string,
	page model.PageRequest,
) ([]*model.Organization, int, error) {
	items := make([]*model.Organization, 0)
	if restricted && len(allowedIDs) == 0 {
		return items, 0, nil
	}
	db := g.DB().Model("organization").Ctx(ctx)
	if restricted {
		db = db.WhereIn("id", allowedIDs)
	}
	if keyword != "" {
		term := "%" + keyword + "%"
		db = db.Where(
			"name LIKE ? OR CAST(org_id AS CHAR) LIKE ? OR description LIKE ?",
			term,
			term,
			term,
		)
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

// CreateOrganization 新建组织。
func (s *Store) CreateOrganization(ctx context.Context, item *model.Organization) (int64, error) {
	result, err := g.DB().Model("organization").Ctx(ctx).Data(item).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateOrganization 使用乐观锁更新组织。
func (s *Store) UpdateOrganization(ctx context.Context, item *model.Organization, version int) (int64, error) {
	type update struct {
		Name        string `db:"name"`
		ParentID    *int64 `db:"parent_id"`
		Description string `db:"description"`
		Status      int    `db:"status"`
		Version     int    `db:"version"`
	}
	result, err := g.DB().Model("organization").Ctx(ctx).
		Where("id", item.ID).Where("version", version).
		Data(update{
			Name: item.Name, ParentID: item.ParentID, Description: item.Description,
			Status: item.Status, Version: version + 1,
		}).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteOrganization 删除无子节点的组织。
func (s *Store) DeleteOrganization(ctx context.Context, id int64) (int64, error) {
	result, err := g.DB().Model("organization").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CountOrganizationChildren 统计直接子组织。
func (s *Store) CountOrganizationChildren(ctx context.Context, id int64) (int, error) {
	return g.DB().Model("organization").Ctx(ctx).Where("parent_id", id).Count()
}

// CountOrganizationUsers 统计组织用户。
func (s *Store) CountOrganizationUsers(ctx context.Context, orgID int64) (int, error) {
	return g.DB().Model("user").Ctx(ctx).Where("org_id", orgID).Count()
}

// CountOrganizationPermissions 统计组织权限。
func (s *Store) CountOrganizationPermissions(ctx context.Context, orgID int64) (int, error) {
	return g.DB().Model("permission").Ctx(ctx).Where("org_id", orgID).Count()
}

// CountOrganizationRoles 统计组织角色。
func (s *Store) CountOrganizationRoles(ctx context.Context, orgID int64) (int, error) {
	return g.DB().Model("role").Ctx(ctx).Where("org_id", orgID).Count()
}

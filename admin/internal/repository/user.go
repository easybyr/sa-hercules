package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *Store) FindUserByUID(ctx context.Context, uid int64) (*model.User, error) {
	var item model.User
	err := g.DB().Model("sys_user").Ctx(ctx).Where("uid", uid).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var item model.User
	err := g.DB().Model("sys_user").Ctx(ctx).Where("username", username).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) ListUsers(ctx context.Context, tenantID int64, keyword string, page model.PageRequest) ([]*model.User, int, error) {
	items := make([]*model.User, 0)
	db := g.DB().Model("sys_user u").Ctx(ctx)
	if tenantID != 0 {
		db = db.InnerJoin("sys_tenant_member tm", "tm.user_id = u.id").Where("tm.tenant_id", tenantID).Where("tm.status", 1)
	}
	if keyword != "" {
		term := "%" + keyword + "%"
		db = db.Where("u.username LIKE ? OR u.nickname LIKE ? OR CAST(u.uid AS CHAR) LIKE ?", term, term, term)
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.Fields("u.*").OrderDesc("u.updated_at").OrderDesc("u.id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

func (s *Store) CreateUser(ctx context.Context, item *model.User, tenantID int64) (int64, error) {
	var userID int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record := struct {
			UID          int64  `db:"uid"`
			Username     string `db:"username"`
			Nickname     string `db:"nickname"`
			PasswordHash string `db:"password_hash"`
			Email        string `db:"email"`
			Phone        string `db:"phone"`
			Status       int    `db:"status"`
			Version      int    `db:"version"`
		}{
			UID:          item.UID,
			Username:     item.Username,
			Nickname:     item.Nickname,
			PasswordHash: item.PasswordHash,
			Email:        item.Email,
			Phone:        item.Phone,
			Status:       item.Status,
			Version:      item.Version,
		}
		result, err := tx.Model("sys_user").Data(record).Insert()
		if err != nil {
			return err
		}
		userID, err = result.LastInsertId()
		if err != nil {
			return err
		}
		_, err = tx.Model("sys_tenant_member").Data(&model.TenantMember{TenantID: tenantID, UserID: userID, Status: 1}).Insert()
		return err
	})
	return userID, err
}

func (s *Store) UserTenantIDs(ctx context.Context, uid int64) ([]int64, error) {
	values, err := g.DB().Model("sys_tenant_member tm").Ctx(ctx).
		InnerJoin("sys_user u", "u.id = tm.user_id").Where("u.uid", uid).
		Where("tm.status", 1).OrderAsc("tm.tenant_id").Array("tm.tenant_id")
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(values))
	for _, value := range values {
		ids = append(ids, value.Int64())
	}
	return ids, nil
}

func (s *Store) UserBelongsToTenant(ctx context.Context, uid int64, tenantID int64) (bool, error) {
	count, err := g.DB().Model("sys_tenant_member tm").Ctx(ctx).
		InnerJoin("sys_user u", "u.id = tm.user_id").Where("u.uid", uid).
		Where("tm.tenant_id", tenantID).Where("tm.status", 1).Count()
	return count > 0, err
}

func (s *Store) TenantMember(ctx context.Context, uid int64, tenantID int64) (*model.TenantMember, error) {
	var item model.TenantMember
	err := g.DB().Model("sys_tenant_member tm").Ctx(ctx).
		InnerJoin("sys_user u", "u.id = tm.user_id").Where("u.uid", uid).
		Where("tm.tenant_id", tenantID).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

func (s *Store) AddUserTenant(ctx context.Context, uid int64, tenantID int64, operations ...*model.PermissionOperationLog) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user model.User
		if err := tx.Model("sys_user").Where("uid", uid).Scan(&user); err != nil {
			return err
		}
		count, err := tx.Model("sys_tenant_member").Where("user_id", user.ID).Where("tenant_id", tenantID).Count()
		if err != nil || count > 0 {
			return err
		}
		if _, err = tx.Model("sys_tenant_member").Data(&model.TenantMember{TenantID: tenantID, UserID: user.ID, Status: 1}).Insert(); err != nil {
			return err
		}
		if len(operations) > 0 && operations[0] != nil {
			return insertPermissionOperationLog(tx, operations[0])
		}
		return nil
	})
}

func (s *Store) UpdateUser(ctx context.Context, item *model.User, version int) (int64, error) {
	record := struct {
		Username     string `db:"username"`
		Nickname     string `db:"nickname"`
		Phone        string `db:"phone"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		Status       int    `db:"status"`
		Version      int    `db:"version"`
	}{item.Username, item.Nickname, item.Phone, item.Email, item.PasswordHash, item.Status, version + 1}
	result, err := g.DB().Model("sys_user").Ctx(ctx).Where("uid", item.UID).Where("version", version).Data(record).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Store) RemoveUserTenant(ctx context.Context, uid, tenantID int64, operation *model.PermissionOperationLog) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user model.User
		if err := tx.Model("sys_user").Where("uid", uid).Scan(&user); err != nil {
			return err
		}
		var member model.TenantMember
		if err := tx.Model("sys_tenant_member").Where("tenant_id", tenantID).Where("user_id", user.ID).Scan(&member); err != nil {
			return err
		}
		if member.ID == 0 {
			return nil
		}
		if _, err := tx.Model("sys_user_role").Where("tenant_id", tenantID).Where("tenant_member_id", member.ID).Delete(); err != nil {
			return err
		}
		var directRole model.Role
		if err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("code", directRoleCode(uid)).Scan(&directRole); err != nil {
			return err
		}
		if directRole.ID != 0 {
			if _, err := tx.Model("sys_role_permission").Where("tenant_id", tenantID).Where("role_id", directRole.ID).Delete(); err != nil {
				return err
			}
			if _, err := tx.Model("sys_role").Where("tenant_id", tenantID).Where("id", directRole.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := tx.Model("sys_tenant_member").Where("id", member.ID).Delete(); err != nil {
			return err
		}
		if operation != nil {
			if err := insertPermissionOperationLog(tx, operation); err != nil {
				return err
			}
		}
		return nil
	})
}

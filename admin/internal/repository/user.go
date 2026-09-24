package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/idgen"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// FindUserByUID 按UID查询用户。
func (s *Store) FindUserByUID(ctx context.Context, uid int64) (*model.User, error) {
	var item model.User
	err := g.DB().Model("user").Ctx(ctx).Where("uid", uid).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// FindUserByUsername 按用户名查询用户。
func (s *Store) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var item model.User
	err := g.DB().Model("user").Ctx(ctx).Where("username", username).Scan(&item)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return &item, err
}

// ListUsers 分页查询用户，orgID为0时查询全部。
func (s *Store) ListUsers(
	ctx context.Context,
	orgID int64,
	keyword string,
	page model.PageRequest,
) ([]*model.User, int, error) {
	items := make([]*model.User, 0)
	db := g.DB().Model("user").Ctx(ctx)
	if orgID != 0 {
		db = db.Where("org_id", orgID)
	}
	if keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	total, err := db.Count()
	if err != nil {
		return nil, 0, err
	}
	err = db.OrderDesc("updated_at").OrderDesc("id").Page(page.Page, page.PageSize).Scan(&items)
	return items, total, err
}

// CreateUser 新建用户。
func (s *Store) CreateUser(ctx context.Context, item *model.User) (int64, error) {
	result, err := g.DB().Model("user").Ctx(ctx).Data(item).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateUser 更新用户。
func (s *Store) UpdateUser(ctx context.Context, item *model.User, version int) (int64, error) {
	type update struct {
		Username string `db:"username"`
		Nickname string `db:"nickname"`
		PhoneNum string `db:"phone_num"`
		Email    string `db:"email"`
		Password string `db:"password"`
		OrgID    int64  `db:"org_id"`
		Status   int    `db:"status"`
		Version  int    `db:"version"`
	}
	result, err := g.DB().Model("user").Ctx(ctx).
		Where("uid", item.UID).Where("version", version).
		Data(update{
			Username: item.Username, Nickname: item.Nickname, PhoneNum: item.PhoneNum,
			Email: item.Email, Password: item.Password, OrgID: item.OrgID,
			Status: item.Status, Version: version + 1,
		}).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteUser 删除用户。
func (s *Store) DeleteUser(ctx context.Context, uid int64) error {
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
		_, err := tx.Model("user").Where("uid", uid).Delete()
		return err
	})
}

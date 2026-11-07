package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

const apiPermissionFields = "id,service_name AS service,http_method AS method,api_path AS path,permission_code,path_type,description,status,created_at,updated_at"

func (s *Store) ListAPIPermissions(ctx context.Context) ([]*model.APIPermission, error) {
	items := make([]*model.APIPermission, 0)
	err := g.DB().Model("sys_api_permission").Ctx(ctx).Fields(apiPermissionFields).
		Where("status", 1).OrderAsc("service_name").OrderAsc("api_path").OrderAsc("http_method").Scan(&items)
	return items, err
}

func (s *Store) EnsureAPIPermission(ctx context.Context, item *model.APIPermission) error {
	count, err := g.DB().Model("sys_api_permission").Ctx(ctx).
		Where("service_name", item.Service).Where("http_method", item.Method).Where("api_path", item.Path).Count()
	if err != nil || count > 0 {
		return err
	}
	_, err = g.DB().Model("sys_api_permission").Ctx(ctx).Data(map[string]any{
		"permission_code": item.PermissionCode, "service_name": item.Service,
		"http_method": item.Method, "api_path": item.Path, "path_type": "EXACT", "status": 1,
	}).Insert()
	return err
}

func (s *Store) APIPermissionCode(ctx context.Context, method, path string) (string, error) {
	value, err := g.DB().Model("sys_api_permission").Ctx(ctx).
		Where("http_method", method).Where("api_path", path).Where("status", 1).Value("permission_code")
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value.String(), err
}

func (s *Store) FindAPIPermissionsByPaths(ctx context.Context, paths []string) ([]*model.APIPermission, error) {
	items := make([]*model.APIPermission, 0)
	if len(paths) == 0 {
		return items, nil
	}
	err := g.DB().Model("sys_api_permission").Ctx(ctx).Fields(apiPermissionFields).
		WhereIn("api_path", paths).Where("status", 1).OrderAsc("api_path").OrderAsc("http_method").Scan(&items)
	return items, err
}

func (s *Store) SavePermissionNode(ctx context.Context, oldCode, code string, nodeType consts.PermissionNodeType, apis []string) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return replacePermissionNodeTx(tx, oldCode, code, nodeType, apis)
	})
}

func ensurePermissionNodeTx(tx gdb.TX, code string, nodeType consts.PermissionNodeType, apis []string) error {
	return replacePermissionNodeTx(tx, code, code, nodeType, apis)
}

func replacePermissionNodeTx(tx gdb.TX, oldCode, code string, _ consts.PermissionNodeType, apis []string) error {
	return replacePermissionAPIsTx(tx, oldCode, code, apis)
}

func replacePermissionAPIsTx(tx gdb.TX, oldCode, code string, paths []string) error {
	if oldCode != "" {
		query := tx.Model("sys_api_permission").Where("permission_code", oldCode)
		if len(paths) > 0 {
			query = query.WhereNotIn("api_path", paths)
		}
		if _, err := query.Data(map[string]any{"permission_code": ""}).Update(); err != nil {
			return err
		}
	}
	if len(paths) == 0 {
		return nil
	}
	_, err := tx.Model("sys_api_permission").WhereIn("api_path", paths).
		Where("(permission_code = '' OR permission_code = ? OR permission_code = ?)", oldCode, code).
		Data(map[string]any{"permission_code": code}).Update()
	return err
}

func (s *Store) PermissionNode(ctx context.Context, code string) (*model.PermissionNode, error) {
	items := make([]*model.APIPermission, 0)
	err := g.DB().Model("sys_api_permission").Ctx(ctx).Fields(apiPermissionFields).
		Where("permission_code", code).Where("status", 1).OrderAsc("api_path").Scan(&items)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(items))
	for _, item := range items {
		paths = append(paths, item.Path)
	}
	encoded, err := json.Marshal(paths)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return &model.PermissionNode{}, nil
	}
	return &model.PermissionNode{ID: 1, PermissionCode: code, APIs: string(encoded)}, nil
}

func (s *Store) FillPermissionAPIs(ctx context.Context, permissions []*model.Permission) error {
	if len(permissions) == 0 {
		return nil
	}
	codes := make([]string, 0, len(permissions))
	seen := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		permission.APIs = []string{}
		if !seen[permission.Code] {
			seen[permission.Code] = true
			codes = append(codes, permission.Code)
		}
	}
	items := make([]*model.APIPermission, 0)
	if err := g.DB().Model("sys_api_permission").Ctx(ctx).Fields(apiPermissionFields).
		WhereIn("permission_code", codes).Where("status", 1).OrderAsc("api_path").Scan(&items); err != nil {
		return err
	}
	byCode := make(map[string][]string)
	for _, item := range items {
		byCode[item.PermissionCode] = append(byCode[item.PermissionCode], item.Path)
	}
	for _, permission := range permissions {
		permission.APIs = byCode[permission.Code]
	}
	return nil
}

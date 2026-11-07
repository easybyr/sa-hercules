package service

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

var permissionCodePattern = regexp.MustCompile(`^(menu|button|api|data)\.[a-z][a-z0-9_]*$`)

// ListPermissions 分页查询权限。
func (s *Service) ListPermissions(
	ctx context.Context,
	claims *token.Claims,
	requestedTenantID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.Permission], error) {
	tenantID, err := listTenantID(claims, requestedTenantID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListPermissions(ctx, tenantID, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	if err = s.store.FillPermissionAPIs(ctx, items); err != nil {
		return nil, err
	}
	return model.NewPageResult(items, total, page), nil
}

// CreatePermission 新建租户权限。
func (s *Service) CreatePermission(
	ctx context.Context,
	claims *token.Claims,
	input *model.PermissionInput,
) (*model.Permission, error) {
	if err := s.RequireTenant(claims, input.TenantID); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(input.Code)
	name := strings.TrimSpace(input.Name)
	apis, err := s.validatePermissionNode(ctx, code, input.NodeType, input.APIs, "")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "PERMISSION_NAME_REQUIRED", "权限名称不能为空")
	}
	tenant, err := s.store.FindTenantByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.ID == 0 {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_TENANT", "租户不存在")
	}
	if err = s.validatePermissionParent(ctx, input.TenantID, 0, input.ParentID); err != nil {
		return nil, err
	}
	existing, err := s.store.FindPermission(ctx, input.TenantID, code)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(http.StatusConflict, "PERMISSION_EXISTS", "租户内权限编码已存在")
	}
	existing, err = s.store.FindPermissionByName(ctx, input.TenantID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(
			http.StatusConflict,
			"PERMISSION_NAME_EXISTS",
			"租户内权限名称已存在，请重新输入权限名称",
		)
	}
	item := &model.Permission{
		Code: code, NodeType: input.NodeType, APIs: apis, Name: name, Description: input.Description,
		TenantID: input.TenantID, ParentID: input.ParentID, Status: 1, Version: 1,
	}
	operation, err := permissionOperationLog(
		claims.UID, input.TenantID, "permission", 0, "permission.create", nil, item,
	)
	if err != nil {
		return nil, err
	}
	item.ID, err = s.store.CreatePermissionWithLog(ctx, item, operation, apis)
	return item, err
}

// UpdatePermission 更新权限。
func (s *Service) UpdatePermission(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	code string,
	input *model.PermissionUpdateInput,
) (*model.Permission, error) {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return nil, err
	}
	item, err := s.store.FindPermission(ctx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, notFound("权限")
	}
	if err = s.validatePermissionParent(ctx, tenantID, item.ID, input.ParentID); err != nil {
		return nil, err
	}
	newCode := strings.TrimSpace(input.Code)
	if newCode == "" {
		newCode = item.Code
	}
	nodeType := input.NodeType
	if nodeType == "" {
		nodeType = item.NodeType
	}
	apis := input.APIs
	if apis == nil {
		apis = item.APIs
	}
	apis, err = s.validatePermissionNode(ctx, newCode, nodeType, apis, item.Code)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "PERMISSION_NAME_REQUIRED", "权限名称不能为空")
	}
	existing, err := s.store.FindPermissionByName(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 && existing.Code != code {
		return nil, apperror.New(
			http.StatusConflict,
			"PERMISSION_NAME_EXISTS",
			"租户内权限名称已存在，请重新输入权限名称",
		)
	}
	if newCode != code {
		existingCode, findErr := s.store.FindPermission(ctx, tenantID, newCode)
		if findErr != nil {
			return nil, findErr
		}
		if existingCode.ID != 0 {
			return nil, apperror.New(http.StatusConflict, "PERMISSION_EXISTS", "新权限编码已被使用")
		}
	}
	after := *item
	after.Code = newCode
	after.NodeType = nodeType
	after.APIs = apis
	after.Name = name
	after.Description = input.Description
	after.ParentID = input.ParentID
	after.Version = input.Version + 1
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "permission", item.ID, "permission.update", item, &after,
	)
	if err != nil {
		return nil, err
	}
	affected, err := s.store.UpdatePermissionWithLog(ctx, tenantID, code, &after, input.Version, operation)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Code = newCode
	item.NodeType = nodeType
	item.APIs = apis
	item.Name = name
	item.Description = input.Description
	item.ParentID = input.ParentID
	item.Version = input.Version + 1
	return item, nil
}

// DeletePermission 删除权限。
func (s *Service) DeletePermission(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	code string,
) error {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return err
	}
	item, err := s.store.FindPermission(ctx, tenantID, code)
	if err != nil {
		return err
	}
	if item.ID == 0 {
		return notFound("权限")
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "permission", item.ID, "permission.delete", item, nil,
	)
	if err != nil {
		return err
	}
	return s.store.DeletePermissionWithLog(ctx, tenantID, code, operation)
}

// ListAPIPermissions 查询已注册 API，供按钮权限选择。
func (s *Service) ListAPIPermissions(ctx context.Context) ([]*model.APIPermission, error) {
	return s.store.ListAPIPermissions(ctx)
}

// PermissionForAPI 查询路由当前绑定的权限编码。
func (s *Service) PermissionForAPI(ctx context.Context, method string, path string) (string, error) {
	return s.store.APIPermissionCode(ctx, method, path)
}

func (s *Service) validatePermissionNode(
	ctx context.Context,
	code string,
	nodeType consts.PermissionNodeType,
	apis []string,
	previousCode string,
) ([]string, error) {
	typedNode := nodeType
	if !typedNode.Valid() {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_PERMISSION_NODE_TYPE", "权限节点类型无效")
	}
	prefix := strings.ToLower(string(typedNode)) + "."
	if !permissionCodePattern.MatchString(code) || !strings.HasPrefix(code, prefix) {
		return nil, apperror.New(
			http.StatusBadRequest,
			"INVALID_PERMISSION_CODE",
			"权限编码前缀必须与节点类型一致，例如 menu.order 或 api.order",
		)
	}
	paths := make([]string, 0, len(apis))
	seen := make(map[string]bool, len(apis))
	for _, path := range apis {
		path = strings.TrimSpace(path)
		if path == "" || !strings.HasPrefix(path, "/") || seen[path] {
			return nil, apperror.New(http.StatusBadRequest, "INVALID_PERMISSION_API", "关联 API 路径无效或重复")
		}
		seen[path] = true
		paths = append(paths, path)
	}
	sort.Strings(paths)
	if typedNode == consts.PermissionNodeTypeMenu && len(paths) > 0 {
		return nil, apperror.New(http.StatusBadRequest, "MENU_PERMISSION_HAS_APIS", "菜单权限不能关联 API")
	}
	if typedNode == consts.PermissionNodeTypeButton && len(paths) == 0 {
		return nil, apperror.New(http.StatusBadRequest, "BUTTON_PERMISSION_API_REQUIRED", "按钮权限必须关联至少一个 API")
	}
	if typedNode == "API" && len(paths) == 0 {
		return nil, apperror.New(http.StatusBadRequest, "API_PERMISSION_API_REQUIRED", "API 权限必须关联至少一个 API")
	}
	registered, err := s.store.FindAPIPermissionsByPaths(ctx, paths)
	if err != nil {
		return nil, err
	}
	registeredPaths := make(map[string]bool, len(registered))
	compatiblePaths := make(map[string]bool, len(registered))
	for _, api := range registered {
		registeredPaths[api.Path] = true
		if api.PermissionCode == "" || api.PermissionCode == code || api.PermissionCode == previousCode {
			compatiblePaths[api.Path] = true
		}
	}
	for _, path := range paths {
		if !registeredPaths[path] {
			return nil, apperror.New(http.StatusBadRequest, "UNKNOWN_PERMISSION_API", "关联 API 未注册")
		}
		if !compatiblePaths[path] {
			return nil, apperror.New(
				http.StatusConflict,
				"PERMISSION_API_ALREADY_ASSIGNED",
				"关联 API 已分配给其他权限",
			)
		}
	}
	return paths, nil
}

func (s *Service) validatePermissionCodes(ctx context.Context, tenantID int64, codes []string) error {
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if code == "" || seen[code] {
			return apperror.New(http.StatusBadRequest, "INVALID_PERMISSION", "权限列表包含空值或重复值")
		}
		seen[code] = true
		permission, err := s.store.FindPermission(ctx, tenantID, code)
		if err != nil {
			return err
		}
		if permission.ID == 0 {
			return apperror.New(http.StatusBadRequest, "INVALID_PERMISSION", "权限不存在或不属于当前租户")
		}
	}
	return nil
}

func (s *Service) validatePermissionParent(ctx context.Context, tenantID, selfID int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	if *parentID == selfID {
		return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "权限不能作为自己的父节点")
	}
	visited := map[int64]bool{selfID: true}
	id := *parentID
	for id != 0 {
		if visited[id] {
			return apperror.New(http.StatusBadRequest, "PERMISSION_CYCLE", "权限层级不能形成循环")
		}
		visited[id] = true
		var parent model.Permission
		if err := s.store.DB().Model("sys_permission").Where("tenant_id", tenantID).Where("id", id).Scan(&parent); err != nil {
			return err
		}
		if parent.ID == 0 {
			return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "父权限不存在或不属于当前租户")
		}
		if parent.ParentID == nil {
			break
		}
		id = *parent.ParentID
	}
	return nil
}

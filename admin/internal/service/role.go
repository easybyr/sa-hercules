package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// ListRoles 查询可见角色，隐藏个人权限角色。
func (s *Service) ListRoles(
	ctx context.Context,
	claims *token.Claims,
	requestedTenantID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.Role], error) {
	tenantID, err := listTenantID(claims, requestedTenantID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListRoles(ctx, tenantID, true, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	return model.NewPageResult(items, total, page), nil
}

// RoleDetail 查询角色及权限。
func (s *Service) RoleDetail(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	roleID int64,
) (map[string]any, error) {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return nil, err
	}
	role, err := s.store.FindRole(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	if role.ID == 0 {
		return nil, notFound("角色")
	}
	codes, err := s.store.RolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"role": role, "permission_codes": codes}, nil
}

// CreateRole 新建角色。
func (s *Service) CreateRole(
	ctx context.Context,
	claims *token.Claims,
	input *model.RoleInput,
) (*model.Role, error) {
	if err := s.RequireTenant(claims, input.TenantID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ROLE_NAME_REQUIRED", "角色名称不能为空")
	}
	tenant, err := s.store.FindTenantByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.ID == 0 {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_TENANT", "租户不存在")
	}
	if err = s.validatePermissionCodes(ctx, input.TenantID, input.PermissionCodes); err != nil {
		return nil, err
	}
	existing, err := s.store.FindRoleByName(ctx, input.TenantID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(
			http.StatusConflict,
			"ROLE_NAME_EXISTS",
			"租户内角色名称已存在，请重新输入角色名称",
		)
	}
	roleCode, err := s.generateRoleCode(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	item := &model.Role{
		Code: roleCode, Name: name, Description: input.Description,
		TenantID: input.TenantID, Status: 1, Version: 1,
	}
	operation, err := permissionOperationLog(
		claims.UID, input.TenantID, "role", 0, "role.create", nil,
		map[string]any{"role": item, "permission_codes": input.PermissionCodes},
	)
	if err != nil {
		return nil, err
	}
	item.ID, err = s.store.CreateRoleWithPermissionsAndLog(ctx, item, input.PermissionCodes, operation)
	item.RoleID = item.ID
	return item, err
}

// UpdateRole 更新角色。
func (s *Service) UpdateRole(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	roleID int64,
	input *model.RoleUpdateInput,
) (*model.Role, error) {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return nil, err
	}
	item, err := s.store.FindRole(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, notFound("角色")
	}
	if item.Code == consts.SuperAdminRoleCode || item.IsSystem != 0 {
		return nil, apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能修改超级管理员角色")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ROLE_NAME_REQUIRED", "角色名称不能为空")
	}
	existing, err := s.store.FindRoleByName(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 && existing.ID != roleID {
		return nil, apperror.New(
			http.StatusConflict,
			"ROLE_NAME_EXISTS",
			"租户内角色名称已存在，请重新输入角色名称",
		)
	}
	previousCodes, err := s.store.RolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	permissionCodes := input.PermissionCodes
	if permissionCodes == nil {
		permissionCodes = previousCodes
	} else if err = s.validatePermissionCodes(ctx, tenantID, permissionCodes); err != nil {
		return nil, err
	}
	after := *item
	after.Name = name
	after.Description = input.Description
	after.Version = input.Version + 1
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "role", roleID, "role.update",
		map[string]any{"role": item, "permission_codes": previousCodes},
		map[string]any{"role": &after, "permission_codes": permissionCodes},
	)
	if err != nil {
		return nil, err
	}
	affected, err := s.store.UpdateRoleWithPermissionsAndLog(
		ctx, tenantID, roleID, name, input.Description, input.Version, permissionCodes, operation,
	)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Name = name
	item.Description = input.Description
	item.Version = input.Version + 1
	return item, nil
}

// DeleteRole 删除角色。
func (s *Service) DeleteRole(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	roleID int64,
) error {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return err
	}
	role, err := s.store.FindRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}
	if role.ID == 0 {
		return notFound("角色")
	}
	if role.Code == consts.SuperAdminRoleCode || role.IsSystem != 0 {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除系统角色")
	}
	codes, err := s.store.RolePermissions(ctx, roleID)
	if err != nil {
		return err
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "role", roleID, "role.delete",
		map[string]any{"role": role, "permission_codes": codes}, nil,
	)
	if err != nil {
		return err
	}
	return s.store.DeleteRoleWithLog(ctx, tenantID, roleID, operation)
}

// AssignRolePermissions 替换角色权限。
func (s *Service) AssignRolePermissions(
	ctx context.Context,
	claims *token.Claims,
	tenantID int64,
	roleID int64,
	codes []string,
) error {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return err
	}
	role, err := s.store.FindRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}
	if role.ID == 0 {
		return notFound("角色")
	}
	if role.IsSystem != 0 {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能修改系统角色")
	}
	if err = s.validatePermissionCodes(ctx, tenantID, codes); err != nil {
		return err
	}
	previousCodes, err := s.store.RolePermissions(ctx, roleID)
	if err != nil {
		return err
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "role", roleID, "role.permissions.replace",
		map[string]any{"permission_codes": previousCodes},
		map[string]any{"permission_codes": codes},
	)
	if err != nil {
		return err
	}
	return s.store.ReplaceRolePermissionsWithLog(ctx, roleID, codes, operation)
}

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
	requestedOrgID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.Role], error) {
	orgID, err := listOrganizationID(claims, requestedOrgID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListRoles(ctx, orgID, false, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	return model.NewPageResult(items, total, page), nil
}

// RoleDetail 查询角色及权限。
func (s *Service) RoleDetail(
	ctx context.Context,
	claims *token.Claims,
	orgID int64,
	roleID int64,
) (map[string]any, error) {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return nil, err
	}
	role, err := s.store.FindRole(ctx, orgID, roleID)
	if err != nil {
		return nil, err
	}
	if role.ID == 0 || role.RoleID <= 0 {
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
	if err := s.RequireOrg(claims, input.OrgID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ROLE_NAME_REQUIRED", "角色名称不能为空")
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, input.OrgID)
	if err != nil {
		return nil, err
	}
	if organization.ID == 0 {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_ORGANIZATION", "组织不存在")
	}
	if err = s.validatePermissionCodes(ctx, input.OrgID, input.PermissionCodes); err != nil {
		return nil, err
	}
	existing, err := s.store.FindRoleByName(ctx, input.OrgID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(
			http.StatusConflict,
			"ROLE_NAME_EXISTS",
			"组织内角色名称已存在，请重新输入角色名称",
		)
	}
	roleID, err := s.generateRoleID(ctx)
	if err != nil {
		return nil, err
	}
	item := &model.Role{
		RoleID: roleID, Name: name, Description: input.Description,
		OrgID: input.OrgID, Version: 1,
	}
	item.ID, err = s.store.CreateRoleWithPermissions(ctx, item, input.PermissionCodes)
	return item, err
}

// UpdateRole 更新角色。
func (s *Service) UpdateRole(
	ctx context.Context,
	claims *token.Claims,
	orgID int64,
	roleID int64,
	input *model.RoleUpdateInput,
) (*model.Role, error) {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return nil, err
	}
	item, err := s.store.FindRole(ctx, orgID, roleID)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 || roleID <= 0 {
		return nil, notFound("角色")
	}
	if roleID == consts.SuperAdminRoleID {
		return nil, apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能修改超级管理员角色")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ROLE_NAME_REQUIRED", "角色名称不能为空")
	}
	existing, err := s.store.FindRoleByName(ctx, orgID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 && existing.RoleID != roleID {
		return nil, apperror.New(
			http.StatusConflict,
			"ROLE_NAME_EXISTS",
			"组织内角色名称已存在，请重新输入角色名称",
		)
	}
	permissionCodes := input.PermissionCodes
	if permissionCodes == nil {
		permissionCodes, err = s.store.RolePermissions(ctx, roleID)
		if err != nil {
			return nil, err
		}
	} else if err = s.validatePermissionCodes(ctx, orgID, permissionCodes); err != nil {
		return nil, err
	}
	affected, err := s.store.UpdateRoleWithPermissions(
		ctx, orgID, roleID, name, input.Description, input.Version, permissionCodes,
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
	orgID int64,
	roleID int64,
) error {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return err
	}
	if roleID == consts.SuperAdminRoleID || roleID <= 0 {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除系统角色")
	}
	role, err := s.store.FindRole(ctx, orgID, roleID)
	if err != nil {
		return err
	}
	if role.ID == 0 {
		return notFound("角色")
	}
	return s.store.DeleteRole(ctx, orgID, roleID)
}

// AssignRolePermissions 替换角色权限。
func (s *Service) AssignRolePermissions(
	ctx context.Context,
	claims *token.Claims,
	orgID int64,
	roleID int64,
	codes []string,
) error {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return err
	}
	role, err := s.store.FindRole(ctx, orgID, roleID)
	if err != nil {
		return err
	}
	if role.ID == 0 || roleID <= 0 {
		return notFound("角色")
	}
	if err = s.validatePermissionCodes(ctx, orgID, codes); err != nil {
		return err
	}
	return s.store.ReplaceRolePermissions(ctx, roleID, codes)
}

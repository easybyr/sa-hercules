package service

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

var permissionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// ListPermissions 分页查询权限。
func (s *Service) ListPermissions(
	ctx context.Context,
	claims *token.Claims,
	requestedOrgID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.Permission], error) {
	orgID, err := listOrganizationID(claims, requestedOrgID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListPermissions(ctx, orgID, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	return model.NewPageResult(items, total, page), nil
}

// CreatePermission 新建组织权限。
func (s *Service) CreatePermission(
	ctx context.Context,
	claims *token.Claims,
	input *model.PermissionInput,
) (*model.Permission, error) {
	if err := s.RequireOrg(claims, input.OrgID); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(input.Code)
	name := strings.TrimSpace(input.Name)
	if !permissionCodePattern.MatchString(code) {
		return nil, apperror.New(
			http.StatusBadRequest,
			"INVALID_PERMISSION_CODE",
			"权限编码格式无效，请使用类似 report.read 的小写点分格式",
		)
	}
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "PERMISSION_NAME_REQUIRED", "权限名称不能为空")
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, input.OrgID)
	if err != nil {
		return nil, err
	}
	if organization.ID == 0 {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_ORGANIZATION", "组织不存在")
	}
	existing, err := s.store.FindPermission(ctx, input.OrgID, code)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(http.StatusConflict, "PERMISSION_EXISTS", "组织内权限编码已存在")
	}
	existing, err = s.store.FindPermissionByName(ctx, input.OrgID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(
			http.StatusConflict,
			"PERMISSION_NAME_EXISTS",
			"组织内权限名称已存在，请重新输入权限名称",
		)
	}
	item := &model.Permission{
		Code: code, Name: name, Description: input.Description,
		OrgID: input.OrgID, Version: 1,
	}
	item.ID, err = s.store.CreatePermission(ctx, item)
	return item, err
}

// UpdatePermission 更新权限。
func (s *Service) UpdatePermission(
	ctx context.Context,
	claims *token.Claims,
	orgID int64,
	code string,
	input *model.PermissionUpdateInput,
) (*model.Permission, error) {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return nil, err
	}
	item, err := s.store.FindPermission(ctx, orgID, code)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, notFound("权限")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "PERMISSION_NAME_REQUIRED", "权限名称不能为空")
	}
	existing, err := s.store.FindPermissionByName(ctx, orgID, name)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 && existing.Code != code {
		return nil, apperror.New(
			http.StatusConflict,
			"PERMISSION_NAME_EXISTS",
			"组织内权限名称已存在，请重新输入权限名称",
		)
	}
	affected, err := s.store.UpdatePermission(
		ctx, orgID, code, name, input.Description, input.Version,
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

// DeletePermission 删除权限。
func (s *Service) DeletePermission(
	ctx context.Context,
	claims *token.Claims,
	orgID int64,
	code string,
) error {
	if err := s.RequireOrg(claims, orgID); err != nil {
		return err
	}
	item, err := s.store.FindPermission(ctx, orgID, code)
	if err != nil {
		return err
	}
	if item.ID == 0 {
		return notFound("权限")
	}
	return s.store.DeletePermission(ctx, orgID, code)
}

func (s *Service) validatePermissionCodes(ctx context.Context, orgID int64, codes []string) error {
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if code == "" || seen[code] {
			return apperror.New(http.StatusBadRequest, "INVALID_PERMISSION", "权限列表包含空值或重复值")
		}
		seen[code] = true
		permission, err := s.store.FindPermission(ctx, orgID, code)
		if err != nil {
			return err
		}
		if permission.ID == 0 {
			return apperror.New(http.StatusBadRequest, "INVALID_PERMISSION", "权限不存在或不属于当前组织")
		}
	}
	return nil
}

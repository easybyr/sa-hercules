package service

import (
	"context"
	"net/http"

	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// Me 获取当前用户完整信息。
func (s *Service) Me(ctx context.Context, claims *token.Claims) (map[string]any, error) {
	user, err := s.store.FindUserByUID(ctx, claims.UID)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, apperror.New(http.StatusNotFound, "USER_NOT_FOUND", "用户不存在")
	}
	roles, err := s.store.UserRoles(ctx, claims.UID, user.OrgID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.store.UserPermissions(ctx, claims.UID, user.OrgID)
	if err != nil {
		return nil, err
	}
	permissionsByCode := make(map[string]*model.Permission, len(permissions))
	for _, permission := range permissions {
		permissionsByCode[permission.Code] = permission
	}
	roleAccesses := make([]*model.RoleAccessView, 0, len(roles))
	for _, role := range roles {
		codes, roleErr := s.store.RolePermissions(ctx, role.RoleID)
		if roleErr != nil {
			return nil, roleErr
		}
		rolePermissions := make([]*model.Permission, 0, len(codes))
		for _, code := range codes {
			if permission, ok := permissionsByCode[code]; ok {
				rolePermissions = append(rolePermissions, permission)
			}
		}
		roleAccesses = append(roleAccesses, &model.RoleAccessView{
			Role:        role,
			Permissions: rolePermissions,
		})
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, user.OrgID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"user": userView(user), "roles": roles,
		"permissions": permissions, "organization": organizationView(organization),
		"role_accesses": roleAccesses, "is_super": claims.IsSuper,
	}, nil
}

// UpdateMe 更新当前用户资料。
func (s *Service) UpdateMe(
	ctx context.Context,
	claims *token.Claims,
	input *model.UserUpdateInput,
) (*model.UserView, error) {
	user, err := s.store.FindUserByUID(ctx, claims.UID)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, apperror.New(http.StatusNotFound, "USER_NOT_FOUND", "用户不存在")
	}
	return s.updateUser(ctx, user, input, false)
}

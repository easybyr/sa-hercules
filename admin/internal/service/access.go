package service

import (
	"context"
	"net/http"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// RequirePermission 校验当前组织内的权限。
func (s *Service) RequirePermission(ctx context.Context, claims *token.Claims, code string) error {
	if claims.IsSuper {
		return nil
	}
	permissions, err := s.store.UserPermissions(ctx, claims.UID, claims.OrgID)
	if err != nil {
		return err
	}
	for _, item := range permissions {
		if item.Code == code {
			return nil
		}
	}
	return apperror.New(http.StatusForbidden, "PERMISSION_DENIED", "没有访问该资源的权限")
}

// ValidateClaims 确保令牌对应用户与组织仍然有效。
func (s *Service) ValidateClaims(ctx context.Context, claims *token.Claims) error {
	user, err := s.store.FindUserByUID(ctx, claims.UID)
	if err != nil {
		return err
	}
	if user.ID == 0 || user.OrgID != claims.OrgID || user.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusUnauthorized, "INVALID_SESSION", "登录状态已失效")
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, user.OrgID)
	if err != nil {
		return err
	}
	if organization.ID == 0 || organization.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusUnauthorized, "INVALID_SESSION", "所属组织不可用")
	}
	return nil
}

// RequireOrg 校验组织隔离。
func (s *Service) RequireOrg(claims *token.Claims, orgID int64) error {
	if claims.IsSuper || claims.OrgID == orgID {
		return nil
	}
	return apperror.New(http.StatusForbidden, "ORGANIZATION_DENIED", "不能访问其他组织的数据")
}

// listOrganizationID 解析列表组织筛选，并保证普通用户只能查询当前组织。
func listOrganizationID(claims *token.Claims, requestedOrgID int64) (int64, error) {
	if claims.IsSuper {
		return requestedOrgID, nil
	}
	if requestedOrgID != 0 && requestedOrgID != claims.OrgID {
		return 0, apperror.New(http.StatusForbidden, "ORGANIZATION_DENIED", "不能访问其他组织的数据")
	}
	return claims.OrgID, nil
}

// CheckAccess 检查权限、角色和组织访问结果。
func (s *Service) CheckAccess(
	ctx context.Context,
	claims *token.Claims,
	permissionCode string,
	roleID int64,
	orgID int64,
) (map[string]bool, error) {
	permissionAllowed := permissionCode == ""
	roleAllowed := roleID == 0
	orgAllowed := orgID == 0 || claims.IsSuper || orgID == claims.OrgID
	if claims.IsSuper {
		permissionAllowed = true
		roleAllowed = true
	} else {
		if permissionCode != "" {
			permissionAllowed = s.RequirePermission(ctx, claims, permissionCode) == nil
		}
		if roleID != 0 {
			roles, err := s.store.UserRoles(ctx, claims.UID, claims.OrgID)
			if err != nil {
				return nil, err
			}
			for _, role := range roles {
				if role.RoleID == roleID {
					roleAllowed = true
					break
				}
			}
		}
	}
	return map[string]bool{
		"permission_allowed":   permissionAllowed,
		"role_allowed":         roleAllowed,
		"organization_allowed": orgAllowed,
		"allowed":              permissionAllowed && roleAllowed && orgAllowed,
	}, nil
}

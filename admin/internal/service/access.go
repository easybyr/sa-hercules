package service

import (
	"context"
	"net/http"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// RequirePermission 校验当前租户内的权限。
func (s *Service) RequirePermission(ctx context.Context, claims *token.Claims, code string) error {
	if claims.IsSuper {
		return nil
	}
	permissions, err := s.store.UserPermissions(ctx, claims.UID, claims.TenantID)
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

// UserHasPermission 判断用户是否属于指定租户并拥有指定权限。
func (s *Service) UserHasPermission(
	ctx context.Context,
	tenantID int64,
	uid int64,
	permissionCode string,
) (bool, error) {
	if tenantID <= 0 || uid <= 0 || permissionCode == "" {
		return false, nil
	}
	return s.store.UserHasPermission(ctx, tenantID, uid, permissionCode)
}

// ValidateClaims 确保令牌对应用户与租户仍然有效。
func (s *Service) ValidateClaims(ctx context.Context, claims *token.Claims) error {
	user, err := s.store.FindUserByUID(ctx, claims.UID)
	if err != nil {
		return err
	}
	if user.ID == 0 || user.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusUnauthorized, "INVALID_SESSION", "登录状态已失效")
	}
	belongs, err := s.store.UserBelongsToTenant(ctx, user.UID, claims.TenantID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperror.New(http.StatusUnauthorized, "INVALID_SESSION", "登录状态已失效")
	}
	tenant, err := s.store.FindTenantByID(ctx, claims.TenantID)
	if err != nil {
		return err
	}
	if tenant.ID == 0 || tenant.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusUnauthorized, "INVALID_SESSION", "所属租户不可用")
	}
	return nil
}

// RequireTenant 校验租户隔离。
func (s *Service) RequireTenant(claims *token.Claims, tenantID int64) error {
	if claims.IsSuper || claims.TenantID == tenantID {
		return nil
	}
	return apperror.New(http.StatusForbidden, "TENANT_DENIED", "不能访问其他租户的数据")
}

// listTenantID 解析列表租户筛选，并保证普通用户只能查询当前租户。
func listTenantID(claims *token.Claims, requestedTenantID int64) (int64, error) {
	if claims.IsSuper {
		return requestedTenantID, nil
	}
	if requestedTenantID != 0 && requestedTenantID != claims.TenantID {
		return 0, apperror.New(http.StatusForbidden, "TENANT_DENIED", "不能访问其他租户的数据")
	}
	return claims.TenantID, nil
}

// CheckAccess 检查权限、角色和租户访问结果。
func (s *Service) CheckAccess(
	ctx context.Context,
	claims *token.Claims,
	permissionCode string,
	roleID int64,
	tenantID int64,
) (map[string]bool, error) {
	permissionAllowed := permissionCode == ""
	roleAllowed := roleID == 0
	tenantAllowed := tenantID == 0 || claims.IsSuper || tenantID == claims.TenantID
	if claims.IsSuper {
		permissionAllowed = true
		roleAllowed = true
	} else {
		if permissionCode != "" {
			permissionAllowed = s.RequirePermission(ctx, claims, permissionCode) == nil
		}
		if roleID != 0 {
			roles, err := s.store.UserRoles(ctx, claims.UID, claims.TenantID)
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
		"permission_allowed": permissionAllowed,
		"role_allowed":       roleAllowed,
		"tenant_allowed":     tenantAllowed,
		"allowed":            permissionAllowed && roleAllowed && tenantAllowed,
	}, nil
}

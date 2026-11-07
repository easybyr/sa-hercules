package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/idgen"
	"github.com/example/sa-hercules/admin/pkg/password"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// ListUsers 查询当前可管理范围内的用户。
func (s *Service) ListUsers(
	ctx context.Context,
	claims *token.Claims,
	requestedTenantID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.UserView], error) {
	tenantID, err := listTenantID(claims, requestedTenantID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListUsers(ctx, tenantID, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	views := make([]*model.UserView, 0, len(items))
	for _, item := range items {
		view, viewErr := s.userView(ctx, item)
		if viewErr != nil {
			return nil, viewErr
		}
		views = append(views, view)
	}
	return model.NewPageResult(views, total, page), nil
}

// UserDetail 查询用户及其租户、角色、权限。
func (s *Service) UserDetail(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	requestedTenantID int64,
) (map[string]any, error) {
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, notFound("用户")
	}
	tenantIDs, err := s.store.UserTenantIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	tenantID := requestedTenantID
	if tenantID == 0 && !claims.IsSuper {
		tenantID = claims.TenantID
	}
	if tenantID == 0 && claims.IsSuper {
		tenantID = firstTenantID(tenantIDs, claims.TenantID)
	}
	if !containsTenantID(tenantIDs, tenantID) {
		return nil, apperror.New(http.StatusForbidden, "TENANT_DENIED", "用户未加入当前租户")
	}
	if err = s.RequireTenant(claims, tenantID); err != nil {
		return nil, err
	}
	roles, err := s.store.UserRoles(ctx, uid, tenantID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.store.UserPermissions(ctx, uid, tenantID)
	if err != nil {
		return nil, err
	}
	directPermissionCodes, err := s.directPermissionCodes(ctx, uid, tenantID)
	if err != nil {
		return nil, err
	}
	tenant, err := s.store.FindTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	view, err := s.userView(ctx, user)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"user":                    view,
		"roles":                   roles,
		"permissions":             permissions,
		"direct_permission_codes": directPermissionCodes,
		"tenant":                  tenantView(tenant),
		"active_tenant_id":        tenantID,
	}, nil
}

// CreateUser 由管理员创建用户。
func (s *Service) CreateUser(
	ctx context.Context,
	claims *token.Claims,
	input *model.UserCreateInput,
) (*model.UserCreateResult, error) {
	status := int(consts.StatusNormal)
	if input.Status != nil {
		status = *input.Status
	}
	if !consts.Status(status).Valid() {
		return nil, invalidStatus()
	}
	if err := s.RequireTenant(claims, input.TenantID); err != nil {
		return nil, err
	}
	initialPassword := input.Password
	if initialPassword == "" {
		var err error
		initialPassword, err = idgen.Password(16)
		if err != nil {
			return nil, err
		}
	} else if len(initialPassword) < 6 {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_PASSWORD", "密码至少需要6位")
	}
	user, err := s.createUserRecord(ctx, &userCreateData{
		Username: input.Username,
		Password: initialPassword,
		Nickname: input.Nickname,
		PhoneNum: input.PhoneNum,
		Email:    input.Email,
		TenantID: input.TenantID,
	}, status)
	if err != nil {
		return nil, err
	}
	return &model.UserCreateResult{User: user, InitialPassword: initialPassword}, nil
}

// UpdateUser 管理员更新用户。
func (s *Service) UpdateUser(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	input *model.UserUpdateInput,
) (*model.UserView, error) {
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, notFound("用户")
	}
	if !claims.IsSuper {
		if err = s.requireUserTenant(ctx, claims, user.UID, claims.TenantID); err != nil {
			return nil, err
		}
	}
	if user.UID == consts.SuperAdminUID && !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "PROTECTED_RESOURCE", "不能修改超级管理员")
	}
	return s.updateUser(ctx, user, input, true)
}

func (s *Service) updateUser(
	ctx context.Context,
	user *model.User,
	input *model.UserUpdateInput,
	allowStatus bool,
) (*model.UserView, error) {
	existing, err := s.store.FindUserByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 && existing.UID != user.UID {
		return nil, apperror.New(http.StatusConflict, "USERNAME_EXISTS", "用户名已存在")
	}
	status := user.Status
	if allowStatus {
		if !consts.Status(input.Status).Valid() {
			return nil, invalidStatus()
		}
		status = input.Status
	}
	hashed := user.PasswordHash
	if input.Password != "" {
		hashed, err = password.Hash(input.Password)
		if err != nil {
			return nil, err
		}
	}
	item := &model.User{
		ID: user.ID, UID: user.UID, Username: strings.TrimSpace(input.Username),
		Nickname: input.Nickname, Phone: input.PhoneNum, Email: input.Email,
		PasswordHash: hashed, Status: status,
	}
	affected, err := s.store.UpdateUser(ctx, item, input.Version)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Version = input.Version + 1
	return s.userView(ctx, item)
}

// DeleteUser 删除用户。
func (s *Service) DeleteUser(ctx context.Context, claims *token.Claims, uid, tenantID int64) error {
	if uid == consts.SuperAdminUID {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除超级管理员")
	}
	if tenantID == 0 {
		tenantID = claims.TenantID
	}
	if err := validateTenantID(tenantID); err != nil {
		return err
	}
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return err
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	belongs, err := s.store.UserBelongsToTenant(ctx, user.UID, tenantID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperror.New(http.StatusForbidden, "TENANT_DENIED", "用户未加入当前租户")
	}
	roles, err := s.store.UserRoles(ctx, uid, tenantID)
	if err != nil {
		return err
	}
	roleIDs := make([]int64, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}
	directCodes, err := s.directPermissionCodes(ctx, uid, tenantID)
	if err != nil {
		return err
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "user", uid, "user.delete",
		map[string]any{"role_ids": roleIDs, "direct_permission_codes": directCodes}, nil,
	)
	if err != nil {
		return err
	}
	return s.store.RemoveUserTenant(ctx, uid, tenantID, operation)
}

// AssignUserRoles 替换用户的普通角色集合。
func (s *Service) AssignUserRoles(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	tenantID int64,
	roleIDs []int64,
) error {
	if tenantID == 0 {
		tenantID = claims.TenantID
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	if err = s.requireUserTenant(ctx, claims, user.UID, tenantID); err != nil {
		return err
	}
	seen := make(map[int64]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID <= 0 || seen[roleID] {
			return apperror.New(http.StatusBadRequest, "INVALID_ROLE", "角色列表包含无效或重复角色")
		}
		seen[roleID] = true
		role, findErr := s.store.FindRole(ctx, tenantID, roleID)
		if findErr != nil {
			return findErr
		}
		if role.ID == 0 {
			return apperror.New(http.StatusBadRequest, "INVALID_ROLE", "角色不存在或不属于当前租户")
		}
	}
	previousRoles, err := s.store.UserRoles(ctx, uid, tenantID)
	if err != nil {
		return err
	}
	previousRoleIDs := make([]int64, 0, len(previousRoles))
	for _, role := range previousRoles {
		previousRoleIDs = append(previousRoleIDs, role.RoleID)
	}
	previousDirectPermissions, err := s.directPermissionCodes(ctx, uid, tenantID)
	if err != nil {
		return err
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "user", uid, "user.roles.replace",
		map[string]any{"role_ids": previousRoleIDs, "direct_permission_codes": previousDirectPermissions},
		map[string]any{"role_ids": roleIDs, "direct_permission_codes": []string{}},
	)
	if err != nil {
		return err
	}
	return s.store.ReplaceUserRolesWithLog(ctx, uid, tenantID, roleIDs, operation)
}

// AssignUserPermissions 通过隐藏的个人角色分配用户直接权限。
func (s *Service) AssignUserPermissions(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	tenantID int64,
	codes []string,
) error {
	if tenantID == 0 {
		tenantID = claims.TenantID
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	if err = s.requireUserTenant(ctx, claims, user.UID, tenantID); err != nil {
		return err
	}
	if err = s.validatePermissionCodes(ctx, tenantID, codes); err != nil {
		return err
	}
	roleID, err := s.ensurePersonalRole(ctx, user, tenantID)
	if err != nil {
		return err
	}
	previousCodes, err := s.store.RolePermissions(ctx, roleID)
	if err != nil {
		return err
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "user", uid, "user.permissions.replace",
		map[string]any{"permission_codes": previousCodes},
		map[string]any{"permission_codes": codes},
	)
	if err != nil {
		return err
	}
	return s.store.ReplaceRolePermissionsWithLog(ctx, roleID, codes, operation)
}

// AssignUserTenant 将用户加入目标租户。
func (s *Service) AssignUserTenant(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	tenantID int64,
) error {
	if !claims.IsSuper {
		return apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "跨租户分配仅限超级管理员")
	}
	if uid == consts.SuperAdminUID {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能修改超级管理员租户")
	}
	tenant, err := s.store.FindTenantByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if tenant.ID == 0 || tenant.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusBadRequest, "INVALID_TENANT", "目标租户不存在或不可用")
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	operation, err := permissionOperationLog(
		claims.UID, tenantID, "user", uid, "user.tenant.add", nil,
		map[string]any{"tenant_id": tenantID},
	)
	if err != nil {
		return err
	}
	return s.store.AddUserTenant(ctx, uid, tenantID, operation)
}

func (s *Service) requireUserTenant(ctx context.Context, claims *token.Claims, uid int64, tenantID int64) error {
	if err := s.RequireTenant(claims, tenantID); err != nil {
		return err
	}
	belongs, err := s.store.UserBelongsToTenant(ctx, uid, tenantID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperror.New(http.StatusForbidden, "TENANT_DENIED", "用户未加入当前租户")
	}
	return nil
}

func containsTenantID(items []int64, tenantID int64) bool {
	for _, item := range items {
		if item == tenantID {
			return true
		}
	}
	return false
}

func firstTenantID(items []int64, fallback int64) int64 {
	if len(items) > 0 {
		return items[0]
	}
	return fallback
}

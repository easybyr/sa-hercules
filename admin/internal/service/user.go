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
	requestedOrgID int64,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.UserView], error) {
	orgID, err := listOrganizationID(claims, requestedOrgID)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListUsers(ctx, orgID, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	views := make([]*model.UserView, 0, len(items))
	for _, item := range items {
		views = append(views, userView(item))
	}
	return model.NewPageResult(views, total, page), nil
}

// UserDetail 查询用户及其组织、角色、权限。
func (s *Service) UserDetail(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
) (map[string]any, error) {
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, notFound("用户")
	}
	if err = s.RequireOrg(claims, user.OrgID); err != nil {
		return nil, err
	}
	roles, err := s.store.UserRoles(ctx, uid, user.OrgID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.store.UserPermissions(ctx, uid, user.OrgID)
	if err != nil {
		return nil, err
	}
	directPermissionCodes, err := s.store.RolePermissions(ctx, -uid)
	if err != nil {
		return nil, err
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, user.OrgID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"user":                    userView(user),
		"roles":                   roles,
		"permissions":             permissions,
		"direct_permission_codes": directPermissionCodes,
		"organization":            organizationView(organization),
	}, nil
}

// CreateUser 由管理员创建用户。
func (s *Service) CreateUser(
	ctx context.Context,
	claims *token.Claims,
	input *model.UserCreateInput,
) (*model.UserCreateResult, error) {
	if input.Status != 0 && !consts.Status(input.Status).Valid() {
		return nil, invalidStatus()
	}
	if err := s.RequireOrg(claims, input.OrgID); err != nil {
		return nil, err
	}
	status := input.Status
	if status == 0 {
		status = int(consts.StatusNormal)
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
		OrgID:    input.OrgID,
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
	if err = s.RequireOrg(claims, user.OrgID); err != nil {
		return nil, err
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
	if allowStatus && input.Status != 0 {
		if !consts.Status(input.Status).Valid() {
			return nil, invalidStatus()
		}
		status = input.Status
	}
	hashed := user.Password
	if input.Password != "" {
		hashed, err = password.Hash(input.Password)
		if err != nil {
			return nil, err
		}
	}
	item := &model.User{
		ID: user.ID, UID: user.UID, Username: strings.TrimSpace(input.Username),
		Nickname: input.Nickname, PhoneNum: input.PhoneNum, Email: input.Email,
		Password: hashed, OrgID: user.OrgID, Status: status,
	}
	affected, err := s.store.UpdateUser(ctx, item, input.Version)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Version = input.Version + 1
	return userView(item), nil
}

// DeleteUser 删除用户。
func (s *Service) DeleteUser(ctx context.Context, claims *token.Claims, uid int64) error {
	if uid == consts.SuperAdminUID {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除超级管理员")
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	if err = s.RequireOrg(claims, user.OrgID); err != nil {
		return err
	}
	return s.store.DeleteUser(ctx, uid)
}

// AssignUserRoles 替换用户的普通角色集合。
func (s *Service) AssignUserRoles(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	roleIDs []int64,
) error {
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	if err = s.RequireOrg(claims, user.OrgID); err != nil {
		return err
	}
	seen := make(map[int64]bool, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID <= 0 || seen[roleID] {
			return apperror.New(http.StatusBadRequest, "INVALID_ROLE", "角色列表包含无效或重复角色")
		}
		seen[roleID] = true
		role, findErr := s.store.FindRole(ctx, user.OrgID, roleID)
		if findErr != nil {
			return findErr
		}
		if role.ID == 0 {
			return apperror.New(http.StatusBadRequest, "INVALID_ROLE", "角色不存在或不属于用户组织")
		}
	}
	return s.store.ReplaceUserRoles(ctx, uid, roleIDs)
}

// AssignUserPermissions 通过隐藏的个人角色分配用户直接权限。
func (s *Service) AssignUserPermissions(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	codes []string,
) error {
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	if err = s.RequireOrg(claims, user.OrgID); err != nil {
		return err
	}
	if err = s.validatePermissionCodes(ctx, user.OrgID, codes); err != nil {
		return err
	}
	roleID, err := s.ensurePersonalRole(ctx, user)
	if err != nil {
		return err
	}
	return s.store.ReplaceRolePermissions(ctx, roleID, codes)
}

// AssignUserOrganization 移动用户到其他组织，并清理原角色关联。
func (s *Service) AssignUserOrganization(
	ctx context.Context,
	claims *token.Claims,
	uid int64,
	orgID int64,
) error {
	if !claims.IsSuper {
		return apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "跨组织分配仅限超级管理员")
	}
	if uid == consts.SuperAdminUID {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能移动超级管理员")
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if organization.ID == 0 || organization.Status != int(consts.StatusNormal) {
		return apperror.New(http.StatusBadRequest, "INVALID_ORGANIZATION", "目标组织不存在或不可用")
	}
	user, err := s.store.FindUserByUID(ctx, uid)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return notFound("用户")
	}
	return s.store.SetUserOrganization(ctx, uid, orgID)
}

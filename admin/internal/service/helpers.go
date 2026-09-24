package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/idgen"
)

func validateUID(uid int64) error {
	if uid <= 0 || uid > 9999999999 {
		return apperror.New(http.StatusBadRequest, "INVALID_UID", "UID必须为不超过10位的正整数")
	}
	return nil
}

func validateOrgID(orgID int64) error {
	if orgID <= 0 || orgID > 99999999 {
		return apperror.New(http.StatusBadRequest, "INVALID_ORG_ID", "组织ID必须为不超过8位的正整数")
	}
	return nil
}

func invalidStatus() error {
	return apperror.New(http.StatusBadRequest, "INVALID_STATUS", "状态值无效")
}

func notFound(resource string) error {
	return apperror.New(http.StatusNotFound, "NOT_FOUND", resource+"不存在")
}

func versionConflict() error {
	return apperror.New(http.StatusConflict, "VERSION_CONFLICT", "数据已被其他请求修改，请刷新后重试")
}

func organizationView(item *model.Organization) *model.OrganizationView {
	return &model.OrganizationView{
		ID: item.ID, OrgID: item.OrgID, Name: item.Name, ParentID: item.ParentID,
		Description: item.Description, Status: consts.Status(item.Status).String(),
		Version: item.Version, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func organizationViews(items []*model.Organization) []*model.OrganizationView {
	views := make([]*model.OrganizationView, 0, len(items))
	for _, item := range items {
		views = append(views, organizationView(item))
	}
	return views
}

func userView(item *model.User) *model.UserView {
	return &model.UserView{
		ID: item.ID, UID: item.UID, Username: item.Username, Nickname: item.Nickname,
		PhoneNum: item.PhoneNum, Email: item.Email, OrgID: item.OrgID,
		Status: consts.Status(item.Status).String(), Version: item.Version,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func (s *Service) ensurePersonalRole(ctx context.Context, user *model.User) (int64, error) {
	roleID := idgen.PersonalRoleID(user.UID)
	role, err := s.store.FindRole(ctx, user.OrgID, roleID)
	if err != nil {
		return 0, err
	}
	if role.ID == 0 {
		role = &model.Role{
			RoleID: roleID, Name: fmt.Sprintf("_user_%d_direct", user.UID),
			Description: "系统维护的用户直接权限角色", OrgID: user.OrgID, Version: 1,
		}
		if _, err = s.store.CreateRole(ctx, role); err != nil {
			return 0, err
		}
	}
	if err = s.store.EnsureUserRole(ctx, user.UID, roleID); err != nil {
		return 0, err
	}
	return roleID, nil
}

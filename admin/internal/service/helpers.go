package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
)

func permissionOperationLog(
	operatorUID int64,
	tenantID int64,
	targetType string,
	targetID int64,
	operation string,
	before any,
	after any,
) (*model.PermissionOperationLog, error) {
	beforeData, err := json.Marshal(before)
	if err != nil {
		return nil, err
	}
	afterData, err := json.Marshal(after)
	if err != nil {
		return nil, err
	}
	return &model.PermissionOperationLog{
		TenantID: tenantID, OperatorUID: operatorUID, TargetType: targetType,
		TargetID: targetID, Operation: operation,
		BeforeData: string(beforeData), AfterData: string(afterData),
	}, nil
}

func validateUID(uid int64) error {
	if uid <= 0 || uid > 9999999999 {
		return apperror.New(http.StatusBadRequest, "INVALID_UID", "UID必须为不超过10位的正整数")
	}
	return nil
}

func validateTenantID(tenantID int64) error {
	if tenantID <= 0 {
		return apperror.New(http.StatusBadRequest, "INVALID_TENANT_ID", "tenant_id必须为正整数")
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

func userView(item *model.User, tenantIDs []int64) *model.UserView {
	return &model.UserView{
		ID: item.ID, UID: item.UID, Username: item.Username, Nickname: item.Nickname,
		PhoneNum: item.Phone, Email: item.Email, TenantIDs: tenantIDs,
		Status: consts.Status(item.Status).String(), Version: item.Version,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func (s *Service) userView(ctx context.Context, item *model.User) (*model.UserView, error) {
	tenantIDs, err := s.store.UserTenantIDs(ctx, item.UID)
	if err != nil {
		return nil, err
	}
	return userView(item, tenantIDs), nil
}

func directRoleCode(uid int64) string { return "user_" + strconv.FormatInt(uid, 10) + "_direct" }

func (s *Service) directPermissionCodes(ctx context.Context, uid, tenantID int64) ([]string, error) {
	role, err := s.store.FindRoleByCode(ctx, tenantID, directRoleCode(uid))
	if err != nil || role.ID == 0 {
		return []string{}, err
	}
	return s.store.RolePermissions(ctx, role.ID)
}

func (s *Service) ensurePersonalRole(ctx context.Context, user *model.User, tenantID int64) (int64, error) {
	role, err := s.store.FindRoleByCode(ctx, tenantID, directRoleCode(user.UID))
	if err != nil {
		return 0, err
	}
	if role.ID == 0 {
		role = &model.Role{
			Code: directRoleCode(user.UID), Name: fmt.Sprintf("_user_%d_direct", user.UID),
			Description: "系统维护的用户直接权限角色", TenantID: tenantID, Status: 1, IsSystem: 1, Version: 1,
		}
		role.ID, err = s.store.CreateRole(ctx, role)
		if err != nil {
			return 0, err
		}
	}
	role.RoleID = role.ID
	if err = s.store.EnsureUserRole(ctx, user.UID, role.ID); err != nil {
		return 0, err
	}
	return role.ID, nil
}

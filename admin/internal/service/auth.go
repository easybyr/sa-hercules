package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/password"
)

func (s *Service) Login(ctx context.Context, input *model.LoginInput) (map[string]any, error) {
	user, err := s.store.FindUserByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		return nil, err
	}
	if user.ID == 0 || !password.Verify(user.PasswordHash, input.Password) {
		return nil, apperror.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
	}
	if consts.Status(user.Status) != consts.StatusNormal {
		return nil, apperror.New(http.StatusForbidden, "USER_DISABLED", "用户已被禁用")
	}
	isSuper := user.UID == consts.SuperAdminUID
	tenantCode := strings.TrimSpace(input.TenantCode)
	if tenantCode == "" && isSuper {
		tenantCode = consts.SuperAdminTenantCode
	}
	if tenantCode == "" {
		return nil, apperror.New(http.StatusBadRequest, "TENANT_CODE_REQUIRED", "请选择登录租户")
	}
	tenant, err := s.store.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		return nil, err
	}
	if tenant.ID == 0 || consts.Status(tenant.Status) != consts.StatusNormal {
		return nil, apperror.New(http.StatusForbidden, "TENANT_DISABLED", "租户不存在或不可用")
	}
	belongs, err := s.store.UserBelongsToTenant(ctx, user.UID, tenant.ID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperror.New(http.StatusForbidden, "TENANT_DENIED", "用户未加入该租户")
	}
	accessToken, err := s.tokens.Issue(user.UID, tenant.ID, user.Username, isSuper)
	if err != nil {
		return nil, err
	}
	view, err := s.userView(ctx, user)
	if err != nil {
		return nil, err
	}
	return map[string]any{"token": accessToken, "token_type": "Bearer", "expires_in": s.tokens.ExpiresIn(), "user": view}, nil
}

func (s *Service) Register(ctx context.Context, input *model.RegisterInput) (*model.UserView, error) {
	tenant, err := s.store.FindTenantByCode(ctx, strings.TrimSpace(input.TenantCode))
	if err != nil {
		return nil, err
	}
	if tenant.ID == 0 || tenant.Status != int(consts.StatusNormal) {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_TENANT", "租户不存在或不可用")
	}
	return s.createUserRecord(ctx, &userCreateData{
		Username: input.Username, Password: input.Password, Nickname: input.Nickname,
		PhoneNum: input.PhoneNum, Email: input.Email, TenantID: tenant.ID,
	}, int(consts.StatusNormal))
}

type userCreateData struct {
	Username string
	Password string
	Nickname string
	PhoneNum string
	Email    string
	TenantID int64
}

func (s *Service) createUserRecord(ctx context.Context, input *userCreateData, status int) (*model.UserView, error) {
	if err := validateTenantID(input.TenantID); err != nil {
		return nil, err
	}
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return nil, apperror.New(http.StatusBadRequest, "USERNAME_REQUIRED", "用户名不能为空")
	}
	existing, err := s.store.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing.ID != 0 {
		return nil, apperror.New(http.StatusConflict, "USERNAME_EXISTS", "用户名已存在")
	}
	tenant, err := s.store.FindTenantByID(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.ID == 0 || tenant.Status != int(consts.StatusNormal) {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_TENANT", "租户不存在或不可用")
	}
	uid, err := s.generateUID(ctx)
	if err != nil {
		return nil, err
	}
	hashed, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	item := &model.User{UID: uid, Username: username, Nickname: strings.TrimSpace(input.Nickname), Phone: input.PhoneNum, Email: input.Email, PasswordHash: hashed, Status: status, Version: 1}
	item.ID, err = s.store.CreateUser(ctx, item, input.TenantID)
	if err != nil {
		return nil, err
	}
	return s.userView(ctx, item)
}

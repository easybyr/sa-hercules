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

// Login 验证用户并签发令牌。
func (s *Service) Login(ctx context.Context, input *model.LoginInput) (map[string]any, error) {
	user, err := s.store.FindUserByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		return nil, err
	}
	if user.ID == 0 || !password.Verify(user.Password, input.Password) {
		return nil, apperror.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误")
	}
	if consts.Status(user.Status) != consts.StatusNormal {
		return nil, apperror.New(http.StatusForbidden, "USER_DISABLED", "用户已被禁用")
	}
	organization, err := s.store.FindOrganizationByOrgID(ctx, user.OrgID)
	if err != nil {
		return nil, err
	}
	if organization.ID == 0 || consts.Status(organization.Status) != consts.StatusNormal {
		return nil, apperror.New(http.StatusForbidden, "ORGANIZATION_DISABLED", "所属组织不可用")
	}
	isSuper := user.UID == consts.SuperAdminUID
	accessToken, err := s.tokens.Issue(user.UID, user.OrgID, user.Username, isSuper)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"token":      accessToken,
		"token_type": "Bearer",
		"expires_in": s.tokens.ExpiresIn(),
		"user":       userView(user),
	}, nil
}

// Register 注册有效组织内的新用户。
func (s *Service) Register(ctx context.Context, input *model.RegisterInput) (*model.UserView, error) {
	return s.createUserRecord(ctx, &userCreateData{
		Username: input.Username,
		Password: input.Password,
		Nickname: input.Nickname,
		PhoneNum: input.PhoneNum,
		Email:    input.Email,
		OrgID:    input.OrgID,
	}, int(consts.StatusNormal))
}

type userCreateData struct {
	Username string
	Password string
	Nickname string
	PhoneNum string
	Email    string
	OrgID    int64
}

func (s *Service) createUserRecord(
	ctx context.Context,
	input *userCreateData,
	status int,
) (*model.UserView, error) {
	if err := validateOrgID(input.OrgID); err != nil {
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
	organization, err := s.store.FindOrganizationByOrgID(ctx, input.OrgID)
	if err != nil {
		return nil, err
	}
	if organization.ID == 0 || consts.Status(organization.Status) != consts.StatusNormal {
		return nil, apperror.New(http.StatusBadRequest, "INVALID_ORGANIZATION", "组织不存在或不可用")
	}
	uid, err := s.generateUID(ctx)
	if err != nil {
		return nil, err
	}
	hashed, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	item := &model.User{
		UID: uid, Username: username, Nickname: strings.TrimSpace(input.Nickname),
		PhoneNum: input.PhoneNum, Email: input.Email, Password: hashed, OrgID: input.OrgID,
		Status: status, Version: 1,
	}
	item.ID, err = s.store.CreateUser(ctx, item)
	if err != nil {
		return nil, err
	}
	return userView(item), nil
}

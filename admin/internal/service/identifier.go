package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/example/sa-hercules/admin/pkg/idgen"
)

const identifierAttempts = 32

func (s *Service) generateUID(ctx context.Context) (int64, error) {
	for attempt := 0; attempt < identifierAttempts; attempt++ {
		uid, err := idgen.Numeric(10)
		if err != nil {
			return 0, err
		}
		user, err := s.store.FindUserByUID(ctx, uid)
		if err != nil {
			return 0, err
		}
		if user.ID == 0 {
			return uid, nil
		}
	}
	return 0, fmt.Errorf("生成唯一UID失败")
}

func (s *Service) generateTenantCode(ctx context.Context) (string, error) {
	for attempt := 0; attempt < identifierAttempts; attempt++ {
		codeValue, err := idgen.Numeric(8)
		if err != nil {
			return "", err
		}
		code := strconv.FormatInt(codeValue, 10)
		tenant, err := s.store.FindTenantByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if tenant.ID == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("生成唯一租户编码失败")
}

func (s *Service) generateRoleCode(ctx context.Context, tenantID int64) (string, error) {
	for attempt := 0; attempt < identifierAttempts; attempt++ {
		roleID, err := idgen.Numeric(10)
		if err != nil {
			return "", err
		}
		code := strconv.FormatInt(roleID, 10)
		role, err := s.store.FindRoleByCode(ctx, tenantID, code)
		if err != nil {
			return "", err
		}
		if role.ID == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("生成唯一角色编码失败")
}

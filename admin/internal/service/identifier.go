package service

import (
	"context"
	"fmt"

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

func (s *Service) generateOrgID(ctx context.Context) (int64, error) {
	for attempt := 0; attempt < identifierAttempts; attempt++ {
		orgID, err := idgen.Numeric(8)
		if err != nil {
			return 0, err
		}
		organization, err := s.store.FindOrganizationByOrgID(ctx, orgID)
		if err != nil {
			return 0, err
		}
		if organization.ID == 0 {
			return orgID, nil
		}
	}
	return 0, fmt.Errorf("生成唯一组织ID失败")
}

func (s *Service) generateRoleID(ctx context.Context) (int64, error) {
	for attempt := 0; attempt < identifierAttempts; attempt++ {
		roleID, err := idgen.Numeric(10)
		if err != nil {
			return 0, err
		}
		role, err := s.store.FindRoleByRoleID(ctx, roleID)
		if err != nil {
			return 0, err
		}
		if role.ID == 0 {
			return roleID, nil
		}
	}
	return 0, fmt.Errorf("生成唯一角色ID失败")
}

package service

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

func (s *Service) SearchTenants(ctx context.Context, keyword string) ([]*model.TenantView, error) {
	term := strings.TrimSpace(keyword)
	if term == "" {
		return []*model.TenantView{}, nil
	}
	items, err := s.store.SearchTenantsExact(ctx, term, 20)
	if err != nil {
		return nil, err
	}
	return tenantViews(items), nil
}

func (s *Service) ListTenants(ctx context.Context, claims *token.Claims, keyword string, page model.PageRequest) (*model.PageResult[*model.TenantView], error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可查看租户")
	}
	items, total, err := s.store.ListTenantsPage(ctx, nil, false, strings.TrimSpace(keyword), page)
	if err != nil {
		return nil, err
	}
	return model.NewPageResult(tenantViews(items), total, page), nil
}

func (s *Service) TenantTree(ctx context.Context, claims *token.Claims, keyword string) ([]*model.TenantView, error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可查看租户")
	}
	items, err := s.store.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	views := tenantViews(items)
	byCode := make(map[string]*model.TenantView, len(views))
	for _, item := range views {
		byCode[item.Code] = item
	}
	roots := make([]*model.TenantView, 0)
	for _, item := range views {
		if item.ParentCode == nil || *item.ParentCode == "" {
			roots = append(roots, item)
			continue
		}
		parent := byCode[*item.ParentCode]
		if parent == nil {
			roots = append(roots, item)
			continue
		}
		parent.Children = append(parent.Children, item)
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Name < roots[j].Name })
	return filterTenantTree(roots, strings.TrimSpace(keyword)), nil
}

func filterTenantTree(items []*model.TenantView, keyword string) []*model.TenantView {
	if keyword == "" {
		return items
	}
	term := strings.ToLower(keyword)
	result := make([]*model.TenantView, 0)
	for _, item := range items {
		children := filterTenantTree(item.Children, keyword)
		matched := strings.Contains(strings.ToLower(item.Name), term) || strings.Contains(strings.ToLower(item.Code), term) || strings.Contains(strings.ToLower(item.Description), term)
		if !matched && len(children) == 0 {
			continue
		}
		view := *item
		view.Children = children
		result = append(result, &view)
	}
	return result
}

func visibleTenants(items []*model.Tenant, tenantID int64) []*model.Tenant {
	byCode := make(map[string]*model.Tenant, len(items))
	children := make(map[string][]*model.Tenant)
	var current *model.Tenant
	for _, item := range items {
		byCode[item.Code] = item
		if item.ID == tenantID {
			current = item
		}
		if item.ParentCode != nil && *item.ParentCode != "" {
			children[*item.ParentCode] = append(children[*item.ParentCode], item)
		}
	}
	if current == nil {
		return nil
	}
	visible := map[string]bool{current.Code: true}
	for parentCode := current.ParentCode; parentCode != nil && *parentCode != ""; {
		parent := byCode[*parentCode]
		if parent == nil || visible[parent.Code] {
			break
		}
		visible[parent.Code] = true
		parentCode = parent.ParentCode
	}
	queue := []*model.Tenant{current}
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		for _, child := range children[item.Code] {
			if !visible[child.Code] {
				visible[child.Code] = true
				queue = append(queue, child)
			}
		}
	}
	result := make([]*model.Tenant, 0, len(visible))
	for _, item := range items {
		if visible[item.Code] {
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) CreateTenant(ctx context.Context, claims *token.Claims, input *model.TenantInput) (*model.TenantView, error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理租户")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "TENANT_NAME_REQUIRED", "租户名称不能为空")
	}
	if err := s.validateTenantParent(ctx, "", input.ParentCode); err != nil {
		return nil, err
	}
	status := int(consts.StatusNormal)
	if input.Status != nil {
		status = *input.Status
	}
	if !consts.Status(status).Valid() {
		return nil, invalidStatus()
	}
	code, err := s.generateTenantCode(ctx)
	if err != nil {
		return nil, err
	}
	item := &model.Tenant{Code: code, Name: name, ParentCode: input.ParentCode, Description: input.Description, Status: status, Version: 1}
	item.ID, err = s.store.CreateTenant(ctx, item)
	if err != nil {
		return nil, err
	}
	return tenantView(item), nil
}

func (s *Service) UpdateTenant(ctx context.Context, claims *token.Claims, id int64, input *model.TenantUpdateInput) (*model.TenantView, error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理租户")
	}
	item, err := s.store.FindTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, notFound("租户")
	}
	if item.Code == consts.SuperAdminTenantCode {
		return nil, apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "系统管理租户不能修改")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "TENANT_NAME_REQUIRED", "租户名称不能为空")
	}
	if !consts.Status(input.Status).Valid() {
		return nil, invalidStatus()
	}
	if err = s.validateTenantParent(ctx, item.Code, input.ParentCode); err != nil {
		return nil, err
	}
	item.Name, item.ParentCode, item.Description, item.Status = name, input.ParentCode, input.Description, input.Status
	affected, err := s.store.UpdateTenant(ctx, item, input.Version)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Version = input.Version + 1
	return tenantView(item), nil
}

func (s *Service) DeleteTenant(ctx context.Context, claims *token.Claims, id int64) error {
	if !claims.IsSuper {
		return apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理租户")
	}
	item, err := s.store.FindTenantByID(ctx, id)
	if err != nil {
		return err
	}
	if item.ID == 0 {
		return notFound("租户")
	}
	if item.Code == consts.SuperAdminTenantCode {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除系统管理租户")
	}
	children, err := s.store.CountTenantChildren(ctx, id)
	if err != nil {
		return err
	}
	if children > 0 {
		return apperror.New(http.StatusConflict, "TENANT_HAS_CHILDREN", "请先删除子租户")
	}
	return s.store.DeleteTenantCascade(ctx, id)
}

func (s *Service) validateTenantParent(ctx context.Context, selfCode string, parentCode *string) error {
	if parentCode == nil || strings.TrimSpace(*parentCode) == "" {
		return nil
	}
	if *parentCode == selfCode {
		return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "租户不能作为自己的上级")
	}
	parent, err := s.store.FindTenantByCode(ctx, *parentCode)
	if err != nil {
		return err
	}
	if parent.ID == 0 {
		return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "上级租户不存在")
	}
	visited := map[string]bool{}
	if selfCode != "" {
		visited[selfCode] = true
	}
	for parent != nil {
		if visited[parent.Code] {
			return apperror.New(http.StatusBadRequest, "TENANT_CYCLE", "租户层级不能形成循环")
		}
		visited[parent.Code] = true
		if parent.ParentCode == nil || *parent.ParentCode == "" {
			break
		}
		parent, err = s.store.FindTenantByCode(ctx, *parent.ParentCode)
		if err != nil {
			return err
		}
		if parent.ID == 0 {
			break
		}
	}
	return nil
}

func tenantView(item *model.Tenant) *model.TenantView {
	return &model.TenantView{ID: item.ID, TenantID: item.ID, Code: item.Code, ParentCode: item.ParentCode, Name: item.Name,
		Status: consts.Status(item.Status).String(), Description: item.Description, Version: item.Version, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func tenantViews(items []*model.Tenant) []*model.TenantView {
	views := make([]*model.TenantView, 0, len(items))
	for _, item := range items {
		views = append(views, tenantView(item))
	}
	return views
}

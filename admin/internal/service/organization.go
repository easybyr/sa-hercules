package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/sa-hercules/admin/internal/consts"
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// SearchOrganizations 搜索注册可选组织。
func (s *Service) SearchOrganizations(ctx context.Context, keyword string) ([]*model.OrganizationView, error) {
	term := strings.TrimSpace(keyword)
	if term == "" {
		return []*model.OrganizationView{}, nil
	}
	items, err := s.store.SearchOrganizationsExact(ctx, term, 20)
	if err != nil {
		return nil, err
	}
	return organizationViews(items), nil
}

// ListOrganizations 分页查询当前用户可见的组织。
func (s *Service) ListOrganizations(
	ctx context.Context,
	claims *token.Claims,
	keyword string,
	page model.PageRequest,
) (*model.PageResult[*model.OrganizationView], error) {
	allowedIDs := make([]int64, 0)
	restricted := !claims.IsSuper
	if restricted {
		allItems, err := s.store.ListOrganizations(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range visibleOrganizations(allItems, claims.OrgID) {
			allowedIDs = append(allowedIDs, item.ID)
		}
	}
	items, total, err := s.store.ListOrganizationsPage(
		ctx,
		allowedIDs,
		restricted,
		strings.TrimSpace(keyword),
		page,
	)
	if err != nil {
		return nil, err
	}
	return model.NewPageResult(organizationViews(items), total, page), nil
}

// OrganizationTree 获取组织树；普通用户仅能看到本组织、上下级组织。
func (s *Service) OrganizationTree(
	ctx context.Context,
	claims *token.Claims,
	keyword string,
) ([]*model.OrganizationView, error) {
	items, err := s.store.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.IsSuper {
		items = visibleOrganizations(items, claims.OrgID)
	}
	views := organizationViews(items)
	byID := make(map[int64]*model.OrganizationView, len(views))
	for _, item := range views {
		byID[item.ID] = item
	}
	roots := make([]*model.OrganizationView, 0)
	for _, item := range views {
		if item.ParentID == nil {
			roots = append(roots, item)
			continue
		}
		parent, ok := byID[*item.ParentID]
		if !ok {
			roots = append(roots, item)
			continue
		}
		parent.Children = append(parent.Children, item)
	}
	return filterOrganizationTree(roots, strings.TrimSpace(keyword)), nil
}

func filterOrganizationTree(
	items []*model.OrganizationView,
	keyword string,
) []*model.OrganizationView {
	if keyword == "" {
		return items
	}
	term := strings.ToLower(keyword)
	result := make([]*model.OrganizationView, 0)
	for _, item := range items {
		children := filterOrganizationTree(item.Children, keyword)
		matched := strings.Contains(strings.ToLower(item.Name), term) ||
			strings.Contains(strconv.FormatInt(item.OrgID, 10), term) ||
			strings.Contains(strings.ToLower(item.Description), term)
		if !matched && len(children) == 0 {
			continue
		}
		view := *item
		view.Children = children
		result = append(result, &view)
	}
	return result
}

func visibleOrganizations(items []*model.Organization, orgID int64) []*model.Organization {
	byID := make(map[int64]*model.Organization, len(items))
	children := make(map[int64][]int64)
	var current *model.Organization
	for _, item := range items {
		byID[item.ID] = item
		if item.OrgID == orgID {
			current = item
		}
		if item.ParentID != nil {
			children[*item.ParentID] = append(children[*item.ParentID], item.ID)
		}
	}
	if current == nil {
		return nil
	}
	visible := map[int64]bool{current.ID: true}
	for parentID := current.ParentID; parentID != nil; {
		parent := byID[*parentID]
		if parent == nil || visible[parent.ID] {
			break
		}
		visible[parent.ID] = true
		parentID = parent.ParentID
	}
	queue := []int64{current.ID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, childID := range children[id] {
			if !visible[childID] {
				visible[childID] = true
				queue = append(queue, childID)
			}
		}
	}
	result := make([]*model.Organization, 0, len(visible))
	for _, item := range items {
		if visible[item.ID] {
			result = append(result, item)
		}
	}
	return result
}

// CreateOrganization 新建组织，仅超级管理员可调用。
func (s *Service) CreateOrganization(
	ctx context.Context,
	claims *token.Claims,
	input *model.OrganizationInput,
) (*model.OrganizationView, error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理组织")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ORGANIZATION_NAME_REQUIRED", "组织名称不能为空")
	}
	if err := s.validateParent(ctx, 0, input.ParentID); err != nil {
		return nil, err
	}
	status := input.Status
	if status == 0 {
		status = int(consts.StatusNormal)
	}
	if !consts.Status(status).Valid() {
		return nil, invalidStatus()
	}
	orgID, err := s.generateOrgID(ctx)
	if err != nil {
		return nil, err
	}
	item := &model.Organization{
		OrgID: orgID, Name: name, ParentID: input.ParentID,
		Description: input.Description, Status: status, Version: 1,
	}
	item.ID, err = s.store.CreateOrganization(ctx, item)
	if err != nil {
		return nil, err
	}
	return organizationView(item), nil
}

// UpdateOrganization 更新组织。
func (s *Service) UpdateOrganization(
	ctx context.Context,
	claims *token.Claims,
	id int64,
	input *model.OrganizationUpdateInput,
) (*model.OrganizationView, error) {
	if !claims.IsSuper {
		return nil, apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理组织")
	}
	item, err := s.store.FindOrganizationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, notFound("组织")
	}
	if item.OrgID == consts.SuperAdminOrgID {
		return nil, apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "系统管理组织不能修改")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperror.New(http.StatusBadRequest, "ORGANIZATION_NAME_REQUIRED", "组织名称不能为空")
	}
	if !consts.Status(input.Status).Valid() {
		return nil, invalidStatus()
	}
	if err = s.validateParent(ctx, id, input.ParentID); err != nil {
		return nil, err
	}
	item.Name = name
	item.ParentID = input.ParentID
	item.Description = input.Description
	item.Status = input.Status
	affected, err := s.store.UpdateOrganization(ctx, item, input.Version)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, versionConflict()
	}
	item.Version = input.Version + 1
	return organizationView(item), nil
}

// DeleteOrganization 删除空组织。
func (s *Service) DeleteOrganization(ctx context.Context, claims *token.Claims, id int64) error {
	if !claims.IsSuper {
		return apperror.New(http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "仅超级管理员可管理组织")
	}
	item, err := s.store.FindOrganizationByID(ctx, id)
	if err != nil {
		return err
	}
	if item.ID == 0 {
		return notFound("组织")
	}
	if item.OrgID == consts.SuperAdminOrgID {
		return apperror.New(http.StatusBadRequest, "PROTECTED_RESOURCE", "不能删除超级管理员组织")
	}
	children, err := s.store.CountOrganizationChildren(ctx, id)
	if err != nil {
		return err
	}
	users, err := s.store.CountOrganizationUsers(ctx, item.OrgID)
	if err != nil {
		return err
	}
	permissions, err := s.store.CountOrganizationPermissions(ctx, item.OrgID)
	if err != nil {
		return err
	}
	roles, err := s.store.CountOrganizationRoles(ctx, item.OrgID)
	if err != nil {
		return err
	}
	if children > 0 || users > 0 || permissions > 0 || roles > 0 {
		return apperror.New(
			http.StatusConflict,
			"ORGANIZATION_NOT_EMPTY",
			"组织包含子组织、用户、角色或权限，不能删除",
		)
	}
	_, err = s.store.DeleteOrganization(ctx, id)
	return err
}

func (s *Service) validateParent(ctx context.Context, selfID int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	if *parentID == selfID {
		return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "组织不能作为自己的上级")
	}
	parent, err := s.store.FindOrganizationByID(ctx, *parentID)
	if err != nil {
		return err
	}
	if parent.ID == 0 {
		return apperror.New(http.StatusBadRequest, "INVALID_PARENT", "上级组织不存在")
	}
	visited := map[int64]bool{selfID: true}
	for parent.ParentID != nil {
		if visited[*parent.ParentID] {
			return apperror.New(http.StatusBadRequest, "ORGANIZATION_CYCLE", "组织层级不能形成循环")
		}
		visited[parent.ID] = true
		parent, err = s.store.FindOrganizationByID(ctx, *parent.ParentID)
		if err != nil {
			return err
		}
		if parent.ID == 0 {
			break
		}
	}
	return nil
}

package service

import (
	"context"
	"testing"

	"github.com/example/sa-hercules/admin/internal/model"
)

func TestSearchOrganizationsRequiresKeyword(t *testing.T) {
	t.Parallel()

	items, err := (&Service{}).SearchOrganizations(context.Background(), "   ")
	if err != nil {
		t.Fatalf("SearchOrganizations() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("SearchOrganizations() returned %d items, want 0", len(items))
	}
}

func TestVisibleOrganizations(t *testing.T) {
	t.Parallel()
	rootID := int64(1)
	currentID := int64(2)
	items := []*model.Organization{
		{ID: rootID, OrgID: 10000000, Name: "root"},
		{ID: currentID, OrgID: 20000000, Name: "current", ParentID: &rootID},
		{ID: 3, OrgID: 30000000, Name: "child", ParentID: &currentID},
		{ID: 4, OrgID: 40000000, Name: "sibling", ParentID: &rootID},
	}

	visible := visibleOrganizations(items, 20000000)
	if len(visible) != 3 {
		t.Fatalf("visible count = %d, want 3", len(visible))
	}
	for _, item := range visible {
		if item.OrgID == 40000000 {
			t.Fatal("sibling organization must not be visible")
		}
	}
}

func TestFilterOrganizationTreeKeepsAncestorPath(t *testing.T) {
	t.Parallel()
	child := &model.OrganizationView{ID: 2, OrgID: 20000000, Name: "研发中心"}
	root := &model.OrganizationView{
		ID: 1, OrgID: 10000000, Name: "系统管理组织",
		Children: []*model.OrganizationView{child},
	}

	filtered := filterOrganizationTree([]*model.OrganizationView{root}, "研发")
	if len(filtered) != 1 || len(filtered[0].Children) != 1 {
		t.Fatalf("filtered tree should preserve the matched node's ancestor path")
	}
	if filtered[0].Children[0].OrgID != child.OrgID {
		t.Fatalf("matched child org_id = %d, want %d", filtered[0].Children[0].OrgID, child.OrgID)
	}
	if len(root.Children) != 1 {
		t.Fatal("filter must not mutate the source tree")
	}
}

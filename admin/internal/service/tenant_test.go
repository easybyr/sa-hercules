package service

import (
	"context"
	"testing"

	"github.com/example/sa-hercules/admin/internal/model"
)

func TestSearchTenantsRequiresKeyword(t *testing.T) {
	t.Parallel()

	items, err := (&Service{}).SearchTenants(context.Background(), "   ")
	if err != nil {
		t.Fatalf("SearchTenants() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("SearchTenants() returned %d items, want 0", len(items))
	}
}

func TestVisibleTenants(t *testing.T) {
	t.Parallel()
	items := []*model.Tenant{
		{ID: 1, Code: "root", Name: "root"},
		{ID: 2, Code: "current", Name: "current", ParentCode: stringPointer("root")},
		{ID: 3, Code: "child", Name: "child", ParentCode: stringPointer("current")},
		{ID: 4, Code: "sibling", Name: "sibling", ParentCode: stringPointer("root")},
	}

	visible := visibleTenants(items, 2)
	if len(visible) != 3 {
		t.Fatalf("visible count = %d, want 3", len(visible))
	}
	for _, item := range visible {
		if item.ID == 4 {
			t.Fatal("sibling tenant must not be visible")
		}
	}
}

func TestFilterTenantTreeKeepsAncestorPath(t *testing.T) {
	t.Parallel()
	child := &model.TenantView{ID: 2, TenantID: 2, Code: "dev", Name: "研发中心"}
	root := &model.TenantView{
		ID: 1, TenantID: 1, Code: "root", Name: "系统管理租户",
		Children: []*model.TenantView{child},
	}

	filtered := filterTenantTree([]*model.TenantView{root}, "研发")
	if len(filtered) != 1 || len(filtered[0].Children) != 1 {
		t.Fatalf("filtered tree should preserve the matched node's ancestor path")
	}
	if filtered[0].Children[0].TenantID != child.TenantID {
		t.Fatalf("matched child tenant_id = %d, want %d", filtered[0].Children[0].TenantID, child.TenantID)
	}
	if len(root.Children) != 1 {
		t.Fatal("filter must not mutate the source tree")
	}
}

func stringPointer(value string) *string { return &value }

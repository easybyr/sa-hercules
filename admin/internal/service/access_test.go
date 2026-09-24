package service

import (
	"testing"

	"github.com/example/sa-hercules/admin/pkg/token"
)

func TestListOrganizationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		claims    *token.Claims
		requested int64
		want      int64
		wantError bool
	}{
		{name: "super all", claims: &token.Claims{IsSuper: true}, want: 0},
		{name: "super selected", claims: &token.Claims{IsSuper: true}, requested: 20000000, want: 20000000},
		{name: "member default", claims: &token.Claims{OrgID: 20000000}, want: 20000000},
		{
			name: "member current", claims: &token.Claims{OrgID: 20000000},
			requested: 20000000, want: 20000000,
		},
		{
			name: "member other", claims: &token.Claims{OrgID: 20000000},
			requested: 30000000, wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := listOrganizationID(test.claims, test.requested)
			if (err != nil) != test.wantError {
				t.Fatalf("listOrganizationID() error = %v, wantError %v", err, test.wantError)
			}
			if actual != test.want {
				t.Fatalf("listOrganizationID() = %d, want %d", actual, test.want)
			}
		})
	}
}

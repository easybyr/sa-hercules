package service

import (
	"testing"

	"github.com/example/sa-hercules/admin/internal/consts"
)

func TestPermissionCodePattern(t *testing.T) {
	valid := []string{"menu.report", consts.PermissionUserAssignRole, "button.report_export3"}
	invalid := []string{"report:read", "report", "Report.read", "menu.", "button.Report"}
	for _, code := range valid {
		if !permissionCodePattern.MatchString(code) {
			t.Errorf("expected %q to be valid", code)
		}
	}
	for _, code := range invalid {
		if permissionCodePattern.MatchString(code) {
			t.Errorf("expected %q to be invalid", code)
		}
	}
}

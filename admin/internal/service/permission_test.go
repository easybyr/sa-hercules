package service

import "testing"

func TestPermissionCodePattern(t *testing.T) {
	valid := []string{"report.read", "user.assign_role", "report2.export3"}
	invalid := []string{"report:read", "report", "Report.read", ".report", "report."}
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

package token

import (
	"testing"
	"time"
)

func TestIssueAndParse(t *testing.T) {
	t.Parallel()
	manager := NewManager("unit-test-secret", time.Hour)
	raw, err := manager.Issue(123, 456, "tester", false)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UID != 123 || claims.OrgID != 456 || claims.Username != "tester" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestExpiredToken(t *testing.T) {
	t.Parallel()
	manager := NewManager("unit-test-secret", -time.Second)
	raw, err := manager.Issue(123, 456, "tester", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Parse(raw); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

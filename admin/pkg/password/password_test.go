package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	t.Parallel()
	hash, err := Hash("safe-password")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(hash, "safe-password") {
		t.Fatal("expected password to verify")
	}
	if Verify(hash, "wrong-password") {
		t.Fatal("expected wrong password to be rejected")
	}
}

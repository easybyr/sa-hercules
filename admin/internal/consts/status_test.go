package consts

import "testing"

func TestStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status Status
		name   string
		valid  bool
	}{
		{status: StatusNormal, name: "normal", valid: true},
		{status: StatusDisabled, name: "disabled", valid: true},
		{status: Status(99), name: "unknown", valid: false},
	}
	for _, test := range tests {
		if got := test.status.String(); got != test.name {
			t.Fatalf("String() = %q, want %q", got, test.name)
		}
		if got := test.status.Valid(); got != test.valid {
			t.Fatalf("Valid() = %v, want %v", got, test.valid)
		}
	}
}

package idgen

import "testing"

func TestNumericLength(t *testing.T) {
	for _, digits := range []int{8, 10} {
		value, err := Numeric(digits)
		if err != nil {
			t.Fatalf("Numeric(%d) returned error: %v", digits, err)
		}
		minimum := int64(1)
		for index := 1; index < digits; index++ {
			minimum *= 10
		}
		if value < minimum || value >= minimum*10 {
			t.Fatalf("Numeric(%d) = %d, length is invalid", digits, value)
		}
	}
}

func TestPasswordLength(t *testing.T) {
	value, err := Password(16)
	if err != nil {
		t.Fatalf("Password returned error: %v", err)
	}
	if len(value) != 16 {
		t.Fatalf("Password length = %d, want 16", len(value))
	}
}

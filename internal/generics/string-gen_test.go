package anbuGenerics

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		wantLength int
	}{
		{"default length", 0, 100},
		{"negative length", -5, 100},
		{"custom length 10", 10, 10},
		{"custom length 45", 45, 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomString(%d) error = %v", tt.length, err)
			}
			if len(got) != tt.wantLength {
				t.Errorf("GenerateRandomString(%d) len = %d, want %d", tt.length, len(got), tt.wantLength)
			}
		})
	}
}

func TestGenerateSequenceString(t *testing.T) {
	seq := GenerateSequenceString(26)
	wantSeq := "abcdefghijklmnopqrstuvwxyz"
	if seq != wantSeq {
		t.Errorf("GenerateSequenceString(26) = %q, want %q", seq, wantSeq)
	}
	if !strings.Contains(seq, "w") {
		t.Errorf("GenerateSequenceString(26) missing letter 'w'")
	}
}

func TestGenerateRUIDString(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		wantLength int
	}{
		{"default for invalid <=0", 0, 18},
		{"default for invalid >30", 35, 18},
		{"valid length 16", 16, 16},
		{"max length 30", 30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRUIDString(tt.length)
			if err != nil {
				t.Fatalf("GenerateRUIDString(%d) error = %v", tt.length, err)
			}
			if len(got) != tt.wantLength {
				t.Errorf("GenerateRUIDString(%d) len = %d, want %d", tt.length, len(got), tt.wantLength)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	t.Run("simple password", func(t *testing.T) {
		pwd, err := GeneratePassword(12, true)
		if err != nil {
			t.Fatalf("GeneratePassword(12, true) error = %v", err)
		}
		if len(pwd) != 12 {
			t.Errorf("len = %d, want 12", len(pwd))
		}
		matched, _ := regexp.MatchString("^[a-z]+$", pwd)
		if !matched {
			t.Errorf("simple password contains non-lowercase chars: %q", pwd)
		}
	})

	t.Run("complex password guarantees character set diversity", func(t *testing.T) {
		pwd, err := GeneratePassword(16, false)
		if err != nil {
			t.Fatalf("GeneratePassword(16, false) error = %v", err)
		}
		if len(pwd) != 16 {
			t.Errorf("len = %d, want 16", len(pwd))
		}
		hasLower := strings.ContainsAny(pwd, lowerSet)
		hasUpper := strings.ContainsAny(pwd, upperSet)
		hasDigit := strings.ContainsAny(pwd, digitSet)
		hasSpecial := strings.ContainsAny(pwd, specialSet)

		if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
			t.Errorf("complex password %q missing set: lower=%v upper=%v digit=%v special=%v", pwd, hasLower, hasUpper, hasDigit, hasSpecial)
		}
	})
}

func TestGeneratePassPhrase(t *testing.T) {
	t.Run("default simple passphrase", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(3, "-", false)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "-")
		if len(parts) != 3 {
			t.Errorf("got %d parts, want 3 in %q", len(parts), phrase)
		}
	})

	t.Run("custom separator and capitalization", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(4, "@", true)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "@")
		if len(parts) != 4 {
			t.Errorf("got %d parts, want 4 in %q", len(parts), phrase)
		}
	})
}

package anbuGenerics

import (
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

func TestGenerateRUIDString(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		wantLength int
	}{
		{"default for invalid <=0", 0, 18},
		{"default for invalid >30", 35, 18},
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

func TestGeneratePassPhrase(t *testing.T) {
	t.Run("clamp below 1", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(0, true)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "-")
		if len(parts) != 3 {
			t.Errorf("got %d parts, want 3 in %q", len(parts), phrase)
		}
	})

	t.Run("clamp above 50", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(51, true)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "-")
		if len(parts) != 3 {
			t.Errorf("got %d parts, want 3 in %q", len(parts), phrase)
		}
	})

	t.Run("simple has no capital and no digit", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(3, true)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "-")
		if len(parts) != 3 {
			t.Errorf("got %d parts, want 3 in %q", len(parts), phrase)
		}
		for _, part := range parts {
			if part == "" {
				t.Fatalf("empty part in %q", phrase)
			}
			for _, r := range part {
				if r >= 'A' && r <= 'Z' {
					t.Errorf("simple phrase %q has a capital in %q", phrase, part)
				}
				if r >= '0' && r <= '9' {
					t.Errorf("simple phrase %q has a digit in %q", phrase, part)
				}
			}
		}
	})

	t.Run("default is one capital and one digit across the phrase", func(t *testing.T) {
		phrase, err := GeneratePassPhrase(3, false)
		if err != nil {
			t.Fatalf("GeneratePassPhrase error = %v", err)
		}
		parts := strings.Split(phrase, "-")
		if len(parts) != 3 {
			t.Errorf("got %d parts, want 3 in %q", len(parts), phrase)
		}
		capCount := 0
		digitCount := 0
		for _, part := range parts {
			if part == "" {
				t.Fatalf("empty part in %q", phrase)
			}
			if part[0] >= 'A' && part[0] <= 'Z' {
				capCount++
			}
			last := part[len(part)-1]
			if last >= '0' && last <= '9' {
				digitCount++
			}
		}
		if capCount != 1 {
			t.Errorf("got %d capitalized words, want 1 in %q", capCount, phrase)
		}
		if digitCount != 1 {
			t.Errorf("got %d words ending in a digit, want 1 in %q", digitCount, phrase)
		}
	})
}

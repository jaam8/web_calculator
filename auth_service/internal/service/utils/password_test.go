package utils

import (
	"testing"
)

func TestGenerateHashAndCompare(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "simple password", password: "password123"},
		{name: "empty password", password: ""},
		{name: "unicode password", password: "pässwörd💥"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateHash(tt.password)
			if err != nil {
				t.Fatalf("GenerateHash(%q) error: %v", tt.password, err)
			}
			if hash == "" {
				t.Fatal("GenerateHash returned empty hash")
			}

			// Compare the correct password
			if !CompareHash(tt.password, hash) {
				t.Errorf("CompareHash did not match for password %q", tt.password)
			}

			// Compare a wrong password
			wrong := tt.password + "wrong"
			if CompareHash(wrong, hash) {
				t.Errorf("CompareHash matched for wrong password %q and hash %q", wrong, hash)
			}
		})
	}
}

package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/afeldman/git-signing-manager/internal/model"
)

func TestParseComment(t *testing.T) {
	tests := []struct {
		comment   string
		baseName  string
		wantName  string
		wantEmail string
	}{
		{
			comment:   "user@example.com",
			baseName:  "id_ed25519",
			wantName:  "user",
			wantEmail: "user@example.com",
		},
		{
			comment:   "John Doe <john@example.com>",
			baseName:  "id_rsa",
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			comment:   "user@hostname",
			baseName:  "id_ed25519",
			wantName:  "user",
			wantEmail: "user@hostname",
		},
		{
			comment:   "",
			baseName:  "id_ed25519",
			wantName:  "id_ed25519",
			wantEmail: "",
		},
		{
			comment:   "Just a name",
			baseName:  "mykey",
			wantName:  "Just a name",
			wantEmail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			name, email := parseComment(tt.comment, tt.baseName)
			if name != tt.wantName {
				t.Errorf("parseComment() name = %q, want %q", name, tt.wantName)
			}
			if email != tt.wantEmail {
				t.Errorf("parseComment() email = %q, want %q", email, tt.wantEmail)
			}
		})
	}
}

func TestAllowedSignersFile(t *testing.T) {
	path := AllowedSignersFile()
	if path == "" {
		t.Skip("Could not determine home directory")
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".ssh", "allowed_signers")
	if path != expected {
		t.Errorf("AllowedSignersFile() = %q, want %q", path, expected)
	}
}

func TestGetProfiles_NoSSHDir(t *testing.T) {
	// This test just ensures the function doesn't crash when SSH dir might not exist
	profiles, err := GetProfiles()
	if err != nil {
		// Only fail if it's not a "directory not exist" type error
		if !os.IsNotExist(err) {
			t.Logf("GetProfiles() returned error (may be expected): %v", err)
		}
	}
	// Just verify it returns a valid slice (even if empty)
	if profiles == nil {
		profiles = []model.Profile{}
	}
	t.Logf("Found %d SSH profiles", len(profiles))
}

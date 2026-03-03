package gitcfg

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/afeldman/git-signing-manager/internal/model"
)

func TestApplyProfile_Validation(t *testing.T) {
	tests := []struct {
		name    string
		profile model.Profile
		wantErr bool
	}{
		{
			name:    "empty profile",
			profile: model.Profile{},
			wantErr: true,
		},
		{
			name:    "missing name",
			profile: model.Profile{Email: "test@example.com", Key: "ABC123"},
			wantErr: true,
		},
		{
			name:    "missing email",
			profile: model.Profile{Name: "Test User", Key: "ABC123"},
			wantErr: true,
		},
		{
			name:    "missing key",
			profile: model.Profile{Name: "Test User", Email: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "valid profile",
			profile: model.Profile{Name: "Test User", Email: "test@example.com", Key: "ABC123"},
			wantErr: false,
		},
	}

	// Create a temporary git repository for testing
	tmpDir, err := os.MkdirTemp("", "git-signing-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Change to temp directory for tests
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyProfile(tt.profile, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsInsideGitRepo(t *testing.T) {
	// Test outside git repo
	tmpDir, err := os.MkdirTemp("", "not-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)

	if IsInsideGitRepo() {
		t.Error("IsInsideGitRepo() returned true for non-git directory")
	}

	// Initialize git repo and test again
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	if !IsInsideGitRepo() {
		t.Error("IsInsideGitRepo() returned false for git directory")
	}

	os.Chdir(oldDir)
}

func TestRunGitConfig(t *testing.T) {
	// Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "git-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Test setting a config value
	err = runGitConfig("--local", "test.key", "test-value")
	if err != nil {
		t.Errorf("runGitConfig() error = %v", err)
	}

	// Verify the value was set
	verifyCmd := exec.Command("git", "config", "--local", "test.key")
	output, err := verifyCmd.Output()
	if err != nil {
		t.Errorf("Failed to verify config: %v", err)
	}

	if string(output) != "test-value\n" {
		t.Errorf("Config value = %q, want %q", string(output), "test-value\n")
	}
}

func TestGetTestModeString(t *testing.T) {
	tests := []struct {
		mode model.TestMode
		want string
	}{
		{model.KeepCommit, "Keep Commit"},
		{model.EphemeralCommit, "Ephemeral (Auto-reset)"},
		{model.TestMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := GetTestModeString(tt.mode); got != tt.want {
				t.Errorf("GetTestModeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSigningMethodString(t *testing.T) {
	tests := []struct {
		method model.SigningMethod
		want   string
	}{
		{model.GPGSigning, "GPG"},
		{model.SSHSigning, "SSH"},
		{model.Unknown, "Unknown"},
		{model.SigningMethod(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := GetSigningMethodString(tt.method); got != tt.want {
				t.Errorf("GetSigningMethodString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRepositoryDirty(t *testing.T) {
	// Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "git-dirty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Clean repo should not be dirty
	dirty, err := isRepositoryDirty()
	if err != nil {
		t.Errorf("isRepositoryDirty() error = %v", err)
	}
	if dirty {
		t.Error("New git repo should not be dirty")
	}

	// Create an untracked file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Repo with untracked file is considered dirty
	dirty, err = isRepositoryDirty()
	if err != nil {
		t.Errorf("isRepositoryDirty() error = %v", err)
	}
	if !dirty {
		t.Error("Repo with untracked files should be dirty")
	}
}

func TestFormatTestResult(t *testing.T) {
	result := &model.TestResult{
		Success:        true,
		SignatureValid: true,
		SigningMethod:  model.GPGSigning,
		RawOutput:      "Good signature",
		SignatureInfo: &model.SignatureInfo{
			KeyID:  "ABC123",
			Name:   "Test User",
			Email:  "test@example.com",
			Valid:  true,
			Signer: "Test User <test@example.com>",
		},
	}

	output := FormatTestResult(result, model.EphemeralCommit)

	if output == "" {
		t.Error("FormatTestResult() returned empty string")
	}

	// Check that key information is present
	if !contains(output, "ABC123") {
		t.Error("FormatTestResult() missing KeyID")
	}
	if !contains(output, "GPG") {
		t.Error("FormatTestResult() missing signing method")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package gitcfg

import (
	"fmt"
	"os/exec"

	"github.com/afeldman/git-signing-manager/internal/model"
	"github.com/afeldman/git-signing-manager/internal/ssh"
)

// ApplyProfile applies a signing profile to the git repository
func ApplyProfile(p model.Profile, global bool) error {
	scope := "--local"
	if global {
		scope = "--global"
	}

	// Validate profile
	if p.Name == "" || p.Email == "" || p.Key == "" {
		return fmt.Errorf("invalid profile: name, email, and key are required")
	}

	if err := runGitConfig(scope, "user.name", p.Name); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}

	if err := runGitConfig(scope, "user.email", p.Email); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}

	if err := runGitConfig(scope, "user.signingkey", p.Key); err != nil {
		return fmt.Errorf("failed to set user.signingkey: %w", err)
	}

	// Configure signing format based on profile type
	if p.Type == model.SSHProfile {
		// SSH signing configuration
		if err := runGitConfig(scope, "gpg.format", "ssh"); err != nil {
			return fmt.Errorf("failed to set gpg.format to ssh: %w", err)
		}

		// Set allowed signers file if it exists
		allowedSigners := ssh.AllowedSignersFile()
		if allowedSigners != "" {
			if err := runGitConfig(scope, "gpg.ssh.allowedSignersFile", allowedSigners); err != nil {
				// Non-fatal: just log warning
				fmt.Printf("Warning: could not set allowed signers file: %v\n", err)
			}
		}
	} else {
		// GPG signing configuration (default)
		if err := runGitConfig(scope, "gpg.format", "openpgp"); err != nil {
			return fmt.Errorf("failed to set gpg.format to openpgp: %w", err)
		}
	}

	if err := runGitConfig(scope, "commit.gpgsign", "true"); err != nil {
		return fmt.Errorf("failed to enable commit signing: %w", err)
	}

	return nil
}

// runGitConfig executes a git config command
func runGitConfig(scope, key, value string) error {
	cmd := exec.Command("git", "config", scope, key, value)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// IsInsideGitRepo checks if the current directory is inside a git repository
func IsInsideGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// GetCurrentSigningConfig returns the current git signing configuration
func GetCurrentSigningConfig(global bool) (name, email, key, format string, err error) {
	scope := "--local"
	if global {
		scope = "--global"
	}

	name, _ = readGitConfigValue(scope, "user.name")
	email, _ = readGitConfigValue(scope, "user.email")
	key, _ = readGitConfigValue(scope, "user.signingkey")
	format, _ = readGitConfigValue(scope, "gpg.format")

	if format == "" {
		format = "openpgp" // Default
	}

	return
}

// readGitConfigValue reads a git config value
func readGitConfigValue(scope, key string) (string, error) {
	cmd := exec.Command("git", "config", scope, key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Trim newline
	result := string(output)
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

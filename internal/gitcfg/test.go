package gitcfg

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/afeldman/git-signing-manager/internal/model"
)

// TestSigning performs a test signing operation with the specified mode
func TestSigning(mode model.TestMode) (*model.TestResult, error) {
	// Check if repository is dirty
	isDirty, err := isRepositoryDirty()
	if err != nil {
		return &model.TestResult{
			Success: false,
			Error:   fmt.Errorf("failed to check repository status: %w", err),
		}, err
	}

	if isDirty {
		// For now, we proceed but mark it in the result
		// In production, you might want to ask for confirmation first
	}

	// Detect signing method
	signingMethod, err := detectSigningMethod()
	if err != nil {
		// Continue with unknown method
		signingMethod = model.Unknown
	}

	// Get current HEAD before test
	currentHead, err := getCurrentHead()
	if err != nil {
		return &model.TestResult{
			Success: false,
			Error:   fmt.Errorf("failed to get current HEAD: %w", err),
		}, err
	}

	// Create signed empty commit with explicit -S flag
	createCmd := exec.Command("git", "commit", "--allow-empty", "-S", "-m", "Test signing")
	var createOut, createErr bytes.Buffer
	createCmd.Stdout = &createOut
	createCmd.Stderr = &createErr

	if err := createCmd.Run(); err != nil {
		// Restore HEAD if ephemeral mode
		if mode == model.EphemeralCommit {
			_ = hardResetToHead(currentHead)
		}
		return &model.TestResult{
			Success:        false,
			SignatureValid: false,
			RawOutput:      createErr.String(),
			Error:          fmt.Errorf("failed to create test commit: %w", err),
			SigningMethod:  signingMethod,
		}, err
	}

	// Get signature output
	logCmd := exec.Command("git", "log", "--show-signature", "-1")
	var logOut bytes.Buffer
	logCmd.Stdout = &logOut
	logCmd.Stderr = &logOut

	if err := logCmd.Run(); err != nil {
		if mode == model.EphemeralCommit {
			_ = hardResetToHead(currentHead)
		}
		return &model.TestResult{
			Success:        false,
			SignatureValid: false,
			RawOutput:      logOut.String(),
			Error:          fmt.Errorf("failed to get signature output: %w", err),
			SigningMethod:  signingMethod,
		}, err
	}

	output := logOut.String()

	// Verify signature
	signatureValid := isSignatureValid(output, signingMethod)

	// Parse signature information
	sigInfo := parseSignatureInfo(output)

	// Check key expiry
	expiryWarning := ""
	if signingMethod == model.GPGSigning && sigInfo != nil && sigInfo.KeyID != "" {
		warning, _ := checkKeyExpiry(sigInfo.KeyID)
		expiryWarning = warning
	}

	// Reset to previous state if ephemeral
	if mode == model.EphemeralCommit {
		if err := hardResetToHead(currentHead); err != nil {
			return &model.TestResult{
				Success:          true,
				SignatureValid:   signatureValid,
				RawOutput:        output,
				Error:            fmt.Errorf("signature valid but failed to reset HEAD: %w", err),
				SigningMethod:    signingMethod,
				KeyExpiryWarning: expiryWarning,
				SignatureInfo:    sigInfo,
			}, err
		}
	}

	return &model.TestResult{
		Success:          true,
		SignatureValid:   signatureValid,
		RawOutput:        output,
		Error:            nil,
		SigningMethod:    signingMethod,
		KeyExpiryWarning: expiryWarning,
		SignatureInfo:    sigInfo,
	}, nil
}

// isRepositoryDirty checks if the working tree has uncommitted changes
func isRepositoryDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// detectSigningMethod detects whether GPG or SSH signing is active
func detectSigningMethod() (model.SigningMethod, error) {
	// Check commit.gpgsign and gpg.format
	localFormat, _ := readGitConfigValue("--local", "gpg.format")
	globalFormat, _ := readGitConfigValue("--global", "gpg.format")

	if strings.TrimSpace(localFormat) == "ssh" || strings.TrimSpace(globalFormat) == "ssh" {
		return model.SSHSigning, nil
	}

	if strings.TrimSpace(localFormat) == "openpgp" || strings.TrimSpace(globalFormat) == "openpgp" {
		return model.GPGSigning, nil
	}

	// Default to GPG if not specified
	return model.GPGSigning, nil
}

// getCurrentHead returns the current HEAD commit hash
func getCurrentHead() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// hardResetToHead performs a hard reset to the specified commit
func hardResetToHead(commitHash string) error {
	if commitHash == "" {
		return fmt.Errorf("invalid commit hash")
	}

	// Verify we only reset one commit by checking the latest message
	latestCmd := exec.Command("git", "log", "-1", "--format=%s")
	latestOut, _ := latestCmd.Output()
	latestMsg := strings.TrimSpace(string(latestOut))

	// Safety check: only reset if the latest commit is our test commit
	if latestMsg != "Test signing" {
		return fmt.Errorf("safety check failed: latest commit is not 'Test signing', refusing to reset")
	}

	cmd := exec.Command("git", "reset", "--hard", commitHash)
	return cmd.Run()
}

// isSignatureValid checks if the signature in the output is valid
func isSignatureValid(output string, method model.SigningMethod) bool {
	// For GPG signatures
	if method == model.GPGSigning || method == model.Unknown {
		if strings.Contains(output, "Good signature") {
			return true
		}
	}

	// For SSH signatures
	if method == model.SSHSigning {
		if strings.Contains(output, "Good \"git\" signature") {
			return true
		}
		if strings.Contains(output, "Valid signature") {
			return true
		}
	}

	return false
}

// parseSignatureInfo extracts signature details from git log output
func parseSignatureInfo(output string) *model.SignatureInfo {
	info := &model.SignatureInfo{}

	// Extract Key ID (GPG)
	// Pattern: "gpg: Signature made ... using RSA key 1234ABCD"
	keyIDRegex := regexp.MustCompile(`(?i)using\s+(?:\w+\s+)?key\s+([A-F0-9]+)`)
	if matches := keyIDRegex.FindStringSubmatch(output); matches != nil {
		info.KeyID = matches[1]
	}

	// Extract signer name and email
	// Pattern: "gpg: Good signature from \"Name <email@example.com>\""
	signerRegex := regexp.MustCompile(`(?:Good signature|Valid signature)\s+(?:from\s+)?["\']([^"\']+)["\']`)
	if matches := signerRegex.FindStringSubmatch(output); matches != nil {
		fullSigner := matches[1]
		info.Signer = fullSigner

		// Try to parse name and email
		angleIdx := strings.Index(fullSigner, "<")
		if angleIdx > 0 && strings.HasSuffix(fullSigner, ">") {
			info.Name = strings.TrimSpace(fullSigner[:angleIdx])
			info.Email = fullSigner[angleIdx+1 : len(fullSigner)-1]
		}
	}

	// Check if signature is valid
	info.Valid = strings.Contains(output, "Good signature") || strings.Contains(output, "Valid signature")

	return info
}

// checkKeyExpiry checks if a GPG key is expiring soon (within 30 days)
func checkKeyExpiry(keyID string) (string, error) {
	// Get GPG key expiry date
	cmd := exec.Command("gpg", "--list-keys", "--with-colons", keyID)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		// pub or sub record
		if parts[0] == "pub" || parts[0] == "sub" {
			// parts[6] is the expiry date (unix timestamp)
			if parts[6] != "" && parts[6] != "0" {
				expiryTime := parseUnixTimestamp(parts[6])
				daysUntilExpiry := time.Until(expiryTime).Hours() / 24

				if daysUntilExpiry < 0 {
					return "⚠️  Key has expired", nil
				}
				if daysUntilExpiry < 30 {
					return fmt.Sprintf("⚠️  Key expires in %.0f days", daysUntilExpiry), nil
				}
			}
			break
		}
	}

	return "", nil
}

// parseUnixTimestamp converts a unix timestamp string to time.Time
func parseUnixTimestamp(timestamp string) time.Time {
	// Try to parse as unix timestamp
	var unixTime int64
	_, err := fmt.Sscanf(timestamp, "%d", &unixTime)
	if err != nil {
		return time.Now()
	}
	return time.Unix(unixTime, 0)
}

// GetTestModeString returns a human-readable string for a TestMode
func GetTestModeString(mode model.TestMode) string {
	switch mode {
	case model.KeepCommit:
		return "Keep Commit"
	case model.EphemeralCommit:
		return "Ephemeral (Auto-reset)"
	default:
		return "Unknown"
	}
}

// GetSigningMethodString returns a human-readable string for a SigningMethod
func GetSigningMethodString(method model.SigningMethod) string {
	switch method {
	case model.GPGSigning:
		return "GPG"
	case model.SSHSigning:
		return "SSH"
	default:
		return "Unknown"
	}
}

// FormatTestResult creates a formatted string representation of TestResult for CLI/UI display
func FormatTestResult(result *model.TestResult, mode model.TestMode) string {
	var output strings.Builder

	// Status line
	if result.SignatureValid {
		output.WriteString("✓ Signature Valid\n")
	} else {
		output.WriteString("✗ Signature Invalid\n")
	}

	// Test mode and signing method
	output.WriteString(fmt.Sprintf("Mode: %s | Method: %s\n", GetTestModeString(mode), GetSigningMethodString(result.SigningMethod)))

	// Signature info
	if result.SignatureInfo != nil && result.SignatureInfo.Valid {
		if result.SignatureInfo.Signer != "" {
			output.WriteString(fmt.Sprintf("Signer: %s\n", result.SignatureInfo.Signer))
		}
		if result.SignatureInfo.KeyID != "" {
			output.WriteString(fmt.Sprintf("Key ID: %s\n", result.SignatureInfo.KeyID))
		}
	}

	// Key expiry warning
	if result.KeyExpiryWarning != "" {
		output.WriteString(fmt.Sprintf("%s\n", result.KeyExpiryWarning))
	}

	// Raw output (if available)
	if result.RawOutput != "" {
		output.WriteString("\n--- Raw Output ---\n")
		output.WriteString(result.RawOutput)
	}

	// Error message
	if result.Error != nil {
		output.WriteString(fmt.Sprintf("\nError: %v\n", result.Error))
	}

	return output.String()
}

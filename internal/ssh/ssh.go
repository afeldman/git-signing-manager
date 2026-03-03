package ssh

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/afeldman/git-signing-manager/internal/model"
)

// AllowedSignersFile returns the default path for the allowed_signers file
func AllowedSignersFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "allowed_signers")
}

// GetProfiles discovers SSH signing keys from the user's .ssh directory
func GetProfiles() ([]model.Profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sshDir := filepath.Join(home, ".ssh")
	var profiles []model.Profile

	// Look for public keys
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil // No SSH directory, return empty
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Look for .pub files (public keys)
		if !strings.HasSuffix(name, ".pub") {
			continue
		}

		// Skip known non-signing keys
		if name == "known_hosts" || name == "authorized_keys" {
			continue
		}

		pubKeyPath := filepath.Join(sshDir, name)
		profile, err := parseSSHPublicKey(pubKeyPath)
		if err != nil {
			continue // Skip invalid keys
		}

		if profile != nil {
			profiles = append(profiles, *profile)
		}
	}

	return profiles, nil
}

// parseSSHPublicKey parses an SSH public key file and extracts profile info
func parseSSHPublicKey(path string) (*model.Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, scanner.Err()
	}

	line := scanner.Text()
	parts := strings.Fields(line)

	if len(parts) < 2 {
		return nil, nil // Invalid key format
	}

	// SSH public key format: <algorithm> <key> [comment]
	// The comment often contains the email
	keyType := parts[0]

	// Only support ed25519 and rsa keys for signing
	if !strings.Contains(keyType, "ssh-ed25519") &&
		!strings.Contains(keyType, "ssh-rsa") &&
		!strings.Contains(keyType, "ecdsa") {
		return nil, nil
	}

	// Extract comment (usually email or user@host)
	comment := ""
	if len(parts) >= 3 {
		comment = strings.Join(parts[2:], " ")
	}

	// Base name without .pub extension
	baseName := strings.TrimSuffix(filepath.Base(path), ".pub")

	// Key path (private key) - remove .pub extension
	keyPath := strings.TrimSuffix(path, ".pub")

	// Try to extract name and email from comment
	name, email := parseComment(comment, baseName)

	return &model.Profile{
		Name:  name,
		Email: email,
		Key:   keyPath, // For SSH signing, we use the private key path
	}, nil
}

// parseComment tries to extract name and email from SSH key comment
func parseComment(comment, baseName string) (name, email string) {
	// Common formats:
	// user@host
	// Name <email@example.com>
	// email@example.com

	if comment == "" {
		return baseName, ""
	}

	// Check for "Name <email>" format
	if strings.Contains(comment, "<") && strings.Contains(comment, ">") {
		start := strings.Index(comment, "<")
		end := strings.Index(comment, ">")
		name = strings.TrimSpace(comment[:start])
		email = comment[start+1 : end]
		return
	}

	// Check for email format
	if strings.Contains(comment, "@") {
		// Could be email or user@host
		parts := strings.Split(comment, "@")
		if len(parts) == 2 {
			// Assume it's an email if the domain part looks like a domain
			if strings.Contains(parts[1], ".") {
				email = comment
				name = parts[0]
			} else {
				// user@host format
				name = parts[0]
				email = comment
			}
		}
		return
	}

	// Just use comment as name
	name = comment
	return
}

// GetSSHSigningKey returns the path to an SSH key suitable for signing
// It checks for ed25519 keys first, then falls back to other types
func GetSSHSigningKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	sshDir := filepath.Join(home, ".ssh")

	// Preferred key types in order
	preferredKeys := []string{
		"id_ed25519",
		"id_ecdsa",
		"id_rsa",
	}

	for _, keyName := range preferredKeys {
		keyPath := filepath.Join(sshDir, keyName)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath, nil
		}
	}

	return "", os.ErrNotExist
}

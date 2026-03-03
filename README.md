# Git Signing Manager

A cross-platform tool for managing GPG and SSH commit signatures in Git. Built with Go, featuring both CLI (BubbleTea) and GUI (Fyne) interfaces.

![Version](https://img.shields.io/badge/version-1.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/go-1.25.6-blue)

## Features

### Core Functionality
- 🔐 **GPG & SSH Signature Support** - Works with both GPG and SSH signing methods
- 👤 **Profile Management** - Save and switch between multiple signing profiles
- 🔄 **Global/Local Configuration** - Apply settings at repository or system level
- 🧪 **Advanced Test Signing** - Comprehensive signature validation with two modes

### Test Signing Modes
- **Ephemeral Mode** - Creates a test commit, verifies the signature, then automatically removes it (leave repo unchanged)
- **Keep Mode** - Creates a test commit with visible signature in git history

### Advanced Features
- ✅ **Automatic Signing Method Detection** - Detects whether GPG or SSH signing is configured
- ⏰ **Key Expiry Warnings** - Alerts when GPG keys are expiring within 30 days
- 🔍 **Signature Parsing** - Extracts and displays key ID, signer name, and email
- 🛡️ **Safety Mechanisms** - Multiple safeguards prevent accidental data loss
- 📊 **Structured Results** - Rich, formatted output with detailed signature information

### User Interfaces
- **CLI (BubbleTea)** - Terminal UI with keyboard shortcuts and color-coded output
- **GUI (Fyne)** - Cross-platform graphical interface with intuitive workflow

## Requirements

- **Go** 1.25.6 or later
- **Git** 2.35.0 or later (for commit signing)
- **GPG** (for GPG signing) or **SSH key** (for SSH signing)
- One of the following:
  - macOS, Linux, or Windows (CLI works everywhere)
  - macOS, Linux, or Windows (GUI supported on all platforms)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/afeldman/git-signing-manager.git
cd git-signing-manager

# Build CLI
go build -o git-signing-manager ./cmd/cli

# Build GUI
go build -o git-signing-manager-gui ./cmd/gui
```

### Using GoReleaser

```bash
# Build all platforms
goreleaser build --clean

# Create and publish release
goreleaser release --clean
```

## Quick Start

### CLI Usage

#### Launch the CLI
```bash
./git-signing-manager
```

#### CLI Keybindings
| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate between profiles |
| `Enter` | Apply selected profile (local scope) |
| `t` | Test signing (ephemeral mode - auto-reset) |
| `T` | Test signing (keep commit mode) |
| `m` | Toggle between test modes |
| `Esc` | Close test result view |
| `q` | Quit |

#### Example CLI Workflow
```
1. Use ↑/↓ to select a signing profile
2. Press Enter to apply the profile locally
3. Press 't' to run a quick signature test (leaves repo unchanged)
4. Review the test result on screen
5. Press Esc to return to profile selection
```

### GUI Usage

#### Launch the GUI
```bash
./git-signing-manager-gui
```

#### GUI Workflow
1. **Select a profile** from the dropdown menu
2. **Optionally check** "Apply globally" to apply system-wide
3. **Click "Apply Profile"** to register the signing configuration
4. **Choose a test mode** from the Test Signing section dropdown
5. **Click "Test Signing"** to validate the configuration
6. **Review the result** in the detailed dialog window

## Architecture

### Package Structure

```
git-signing-manager/
├── cmd/
│   ├── cli/
│   │   ├── main.go          # CLI entry point
│   │   └── model.go         # BubbleTea UI model
│   └── gui/
│       └── main.go          # GUI entry point (Fyne)
├── internal/
│   ├── gitcfg/
│   │   ├── git.go           # Git configuration management
│   │   └── test.go          # Advanced test signing implementation
│   ├── gpg/
│   │   └── gpg.go           # GPG profile discovery
│   └── model/
│       └── profile.go       # Type definitions & domain models
└── go.mod                    # Go module definition
```

### Core Types

```go
// Test signing modes
type TestMode int
const (
    KeepCommit      // Creates signed commit, keeps in history
    EphemeralCommit // Creates signed commit, auto-resets
)

// Detected signing methods
type SigningMethod int
const (
    GPGSigning  // GPG signature method
    SSHSigning  // SSH signature method
    Unknown     // Could not detect
)

// Parsed signature information
type SignatureInfo struct {
    KeyID  string  // GPG key ID
    Name   string  // Signer name
    Email  string  // Signer email
    Valid  bool    // Is signature valid
    Signer string  // Full signer string
}

// Test result with comprehensive details
type TestResult struct {
    Success          bool           // Operation succeeded
    SignatureValid   bool           // Signature is valid
    RawOutput        string         // Full git output
    Error            error          // Any error encountered
    SigningMethod    SigningMethod  // Detected signing method
    KeyExpiryWarning string         // Key expiry warning (if any)
    SignatureInfo    *SignatureInfo // Parsed signature details
}
```

### Key Functions

#### Profile Management
```go
// Apply a profile to git configuration
gitcfg.ApplyProfile(profile, global bool) error

// Discover GPG profiles from system
gpg.GetProfiles() ([]Profile, error)
```

#### Test Signing (Advanced)
```go
// Test signing with structured result
gitcfg.TestSigning(mode TestMode) (*TestResult, error)

// Helper functions
gitcfg.GetTestModeString(mode) string
gitcfg.GetSigningMethodString(method) string
gitcfg.FormatTestResult(result, mode) string
```

## Configuration

### Git Configuration Scope

The tool respects Git's configuration hierarchy:

```bash
# Local scope (per-repository)
git config --local user.name "Your Name"
git config --local user.email "your.email@example.com"
git config --local user.signingkey "KEY_ID"
git config --local commit.gpgsign true

# Global scope (system-wide)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
git config --global user.signingkey "KEY_ID"
git config --global commit.gpgsign true
```

### Signing Method Configuration

```bash
# Use GPG (default)
git config gpg.format openpgp

# Use SSH
git config gpg.format ssh

# Specify SSH key path
git config user.signingkey "/path/to/ssh/key.pub"
```

## Safety Features

### Built-in Safeguards

- ✅ **Repository Dirty Check** - Warns before operations on uncommitted changes
- ✅ **HEAD Preservation** - Saves current HEAD before any modifications
- ✅ **Commit Message Verification** - Only resets commits marked as "Test signing"
- ✅ **Single-Commit Reset Guarantee** - Never resets more than one commit
- ✅ **Error Recovery** - Automatic rollback on failures

### Preventing Accidental Data Loss

The ephemeral test mode includes multiple safety checks:

1. Verifies repository status before creating test commit
2. Saves HEAD commit hash before modification
3. Verifies commit message equals "Test signing" before reset
4. Uses `git reset --hard` only after all safety checks pass
5. Rolls back changes if signature validation fails

## Examples

### Example 1: CLI Test Signing

```
$ ./git-signing-manager

Git Signing Manager (CLI)
═══════════════════════════════════════

Available Profiles:

> John Doe <john@example.com>
  Jane Smith <jane@example.com>

Test Mode: Ephemeral (Auto-reset)

[✓] Profile applied: John Doe <john@example.com>

Commands:
  ↑/↓     navigate profiles
  enter   apply selected profile
  t       test signing (ephemeral)
  T       test signing (keep commit)
  m       toggle test mode
  q       quit

# Press 't' for ephemeral test
# Press 'T' for keep test
```

### Example 2: Test Result (CLI)

```
Test Signing Result
═══════════════════════════════════════

✓ Signature Valid
Mode: Ephemeral (Auto-reset)
Method: GPG
Signer: John Doe <john@example.com>
Key ID: 1234ABCDEF567890

--- Raw Output (selected lines) ---
commit abc123def456
Author: John Doe <john@example.com>
Date:   Wed Mar 3 14:59:00 2026 +0100

    Test signing

gpg: Signature made Wed Mar 3 14:59:00 2026 +0100
gpg: Good signature from "John Doe <john@example.com>"
```

### Example 3: GUI Test Signing Flow

```
1. Launches GUI window
2. User selects "John Doe <john@example.com>" from dropdown
3. Selects test mode "Ephemeral (Auto-reset)"
4. Clicks "Test Signing" button
5. Result dialog appears with:
   ✓ Signature Valid
   Mode: Ephemeral (Auto-reset)
   Method: GPG
   Signer: John Doe <john@example.com>
   Key ID: 1234ABCDEF567890
   [... expandable raw output ...]
```

## Advanced Usage

### Detecting Signing Method

The tool automatically detects your signing method:

```bash
# Check detected method
# Application will show "GPG" or "SSH" in test results
```

### Key Expiry Warnings

The tool alerts you when your GPG key is expiring:

```
⚠️  Key expires in 14 days

# Shows automatically in:
# - CLI test result view
# - GUI result dialog
```

### Custom Profiles

Profiles are discovered from your GPG keyring:

```bash
# List your available keys
gpg --list-secret-keys

# Each key creates an available profile with:
# - Key ID
# - Name
# - Email
```

## Troubleshooting

### Issue: "Failed to create test commit"

**Cause**: Git signing not properly configured  
**Solution**: 
```bash
# Verify GPG configuration
git config user.signingkey
git config gpg.format

# Test GPG directly
echo "test" | gpg --detach-sign
```

### Issue: "Signature Invalid"

**Cause**: GPG/SSH key not loaded or expired  
**Solution**:
```bash
# For GPG, ensure agent is running
gpg-agent --version

# For SSH, add key to agent
ssh-add ~/.ssh/id_ed25519
```

### Issue: "Key has expired"

**Cause**: GPG key expiry date has passed  
**Solution**:
```bash
# Extend key expiry
gpg --edit-key <KEY_ID>
# Follow interactive prompts
```

### Issue: "Hard reset disabled"

**Cause**: Commit message doesn't match "Test signing"  
**Solution**: This is a safety feature. Only ephemeral test commits can trigger auto-reset.

## Building from Source

### Prerequisites

```bash
# Install Go 1.25.6 or later
go version

# Verify Git is installed
git --version
```

### Build Steps

```bash
# Clone repository
git clone https://github.com/afeldman/git-signing-manager.git
cd git-signing-manager

# Download dependencies
go mod download

# Build CLI
go build -o bin/git-signing-manager ./cmd/cli

# Build GUI
go build -o bin/git-signing-manager-gui ./cmd/gui

# Run CLI
./bin/git-signing-manager

# Run GUI
./bin/git-signing-manager-gui
```

### Build with GoReleaser

```bash
# Install GoReleaser
brew install goreleaser

# Build for current platform
goreleaser build --snapshot --clean

# Build for all platforms
goreleaser build --clean

# Create release
goreleaser release --clean
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Style

The project follows standard Go conventions:

```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run linter (if installed)
golangci-lint run
```

## Dependencies

### Core Dependencies

- **fyne.io/fyne/v2** (v2.7.3+) - GUI framework
- **charm.land/bubbletea** (v2.0.0+) - CLI framework (TUI)

### External Tools (not included)

- **git** - Version control (must be installed)
- **gpg** - GPG signing support (optional)
- **ssh-agent** - SSH signing support (optional)

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or suggestions:

1. Check existing [GitHub Issues](https://github.com/afeldman/git-signing-manager/issues)
2. Review the [Feature Guide](FEATURE_GUIDE.md) and [Implementation Summary](IMPLEMENTATION_SUMMARY.md)
3. Create a new issue with detailed information

## Roadmap

### Planned Features

- [ ] Async test signing operations
- [ ] Configuration file support (YAML/TOML)
- [ ] Statistics tracking (key usage, signature counts)
- [ ] Email notifications for key expiry
- [ ] Web-based management interface
- [ ] Integration with password managers
- [ ] Signature verification for existing commits
- [ ] Batch operations support

## Changelog

### Version 1.0 (Current)

**Features:**
- Profile selection and management
- Global/Local configuration switching
- Advanced test signing (two modes)
- Signing method detection (GPG/SSH)
- Signature parsing and validation
- Key expiry warnings
- CLI interface with BubbleTea
- GUI interface with Fyne
- Safety mechanisms to prevent data loss
- Comprehensive error handling

## Authors

- **Anton Feldmann** - Initial implementation

## Acknowledgments

- [BubbleTea](https://github.com/charmbracelet/bubbletea) - Great terminal UI framework
- [Fyne](https://fyne.io) - Excellent cross-platform GUI toolkit
- [Git](https://git-scm.com/) - The best version control system

---

**Last Updated**: March 3, 2026  
**Status**: Production Ready ✅

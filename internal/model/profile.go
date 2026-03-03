package model

// TestMode defines how the test commit should be handled
type TestMode int

const (
	// KeepCommit mode creates a signed commit and keeps it in history
	KeepCommit TestMode = iota
	// EphemeralCommit mode creates a signed commit, verifies it, then resets HEAD
	EphemeralCommit
)

// SigningMethod indicates which signing method is active
type SigningMethod int

const (
	GPGSigning SigningMethod = iota
	SSHSigning
	Unknown
)

// SignatureInfo contains parsed signature details
type SignatureInfo struct {
	KeyID  string
	Name   string
	Email  string
	Valid  bool
	Signer string
}

// TestResult contains the structured result of a test signing operation
type TestResult struct {
	Success          bool
	SignatureValid   bool
	RawOutput        string
	Error            error
	SigningMethod    SigningMethod
	KeyExpiryWarning string // Warning message if key expires soon
	SignatureInfo    *SignatureInfo
}

// ProfileType indicates the type of signing key
type ProfileType int

const (
	GPGProfile ProfileType = iota
	SSHProfile
)

// String returns a human-readable string for ProfileType
func (pt ProfileType) String() string {
	switch pt {
	case GPGProfile:
		return "GPG"
	case SSHProfile:
		return "SSH"
	default:
		return "Unknown"
	}
}

// Profile represents a signing profile (GPG or SSH)
type Profile struct {
	Name  string
	Email string
	Key   string
	Type  ProfileType // GPG or SSH
}

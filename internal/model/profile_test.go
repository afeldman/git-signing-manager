package model

import "testing"

func TestTestMode_Constants(t *testing.T) {
	// Ensure constants have expected values
	if KeepCommit != 0 {
		t.Errorf("KeepCommit = %d, want 0", KeepCommit)
	}
	if EphemeralCommit != 1 {
		t.Errorf("EphemeralCommit = %d, want 1", EphemeralCommit)
	}
}

func TestSigningMethod_Constants(t *testing.T) {
	// Ensure constants have expected values
	if GPGSigning != 0 {
		t.Errorf("GPGSigning = %d, want 0", GPGSigning)
	}
	if SSHSigning != 1 {
		t.Errorf("SSHSigning = %d, want 1", SSHSigning)
	}
	if Unknown != 2 {
		t.Errorf("Unknown = %d, want 2", Unknown)
	}
}

func TestProfile_Fields(t *testing.T) {
	p := Profile{
		Name:  "Test User",
		Email: "test@example.com",
		Key:   "ABC123DEF456",
	}

	if p.Name != "Test User" {
		t.Errorf("Profile.Name = %q, want %q", p.Name, "Test User")
	}
	if p.Email != "test@example.com" {
		t.Errorf("Profile.Email = %q, want %q", p.Email, "test@example.com")
	}
	if p.Key != "ABC123DEF456" {
		t.Errorf("Profile.Key = %q, want %q", p.Key, "ABC123DEF456")
	}
}

func TestSignatureInfo_Fields(t *testing.T) {
	sig := SignatureInfo{
		KeyID:  "ABC123",
		Name:   "Test User",
		Email:  "test@example.com",
		Valid:  true,
		Signer: "Test User <test@example.com>",
	}

	if sig.KeyID != "ABC123" {
		t.Errorf("SignatureInfo.KeyID = %q, want %q", sig.KeyID, "ABC123")
	}
	if !sig.Valid {
		t.Error("SignatureInfo.Valid = false, want true")
	}
}

func TestTestResult_Fields(t *testing.T) {
	result := TestResult{
		Success:          true,
		SignatureValid:   true,
		RawOutput:        "test output",
		SigningMethod:    GPGSigning,
		KeyExpiryWarning: "",
		SignatureInfo:    nil,
		Error:            nil,
	}

	if !result.Success {
		t.Error("TestResult.Success = false, want true")
	}
	if !result.SignatureValid {
		t.Error("TestResult.SignatureValid = false, want true")
	}
	if result.SigningMethod != GPGSigning {
		t.Errorf("TestResult.SigningMethod = %d, want %d", result.SigningMethod, GPGSigning)
	}
}

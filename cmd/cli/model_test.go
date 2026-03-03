package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afeldman/git-signing-manager/internal/model"
)

// Helper to create a key press message
func keyPress(key rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

// Helper to create a special key press (arrow keys, etc.)
func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestColorizeStatus(t *testing.T) {
	tests := []struct {
		color string
		want  string
	}{
		{"green", "✓"},
		{"red", "✗"},
		{"yellow", "⚠"},
		{"unknown", "•"},
		{"", "•"},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			if got := colorizeStatus(tt.color); got != tt.want {
				t.Errorf("colorizeStatus(%q) = %q, want %q", tt.color, got, tt.want)
			}
		})
	}
}

func TestCliModel_Navigation(t *testing.T) {
	m := cliModel{
		profiles: []model.Profile{
			{Name: "User 1", Email: "user1@example.com", Key: "KEY1", Type: model.GPGProfile},
			{Name: "User 2", Email: "user2@example.com", Key: "KEY2", Type: model.GPGProfile},
			{Name: "User 3", Email: "user3@example.com", Key: "KEY3", Type: model.SSHProfile},
		},
		cursor:   0,
		testMode: model.EphemeralCommit,
	}

	// Test down navigation
	updatedModel, _ := m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(cliModel)
	if m.cursor != 1 {
		t.Errorf("cursor after 'down' = %d, want 1", m.cursor)
	}

	// Test down again
	updatedModel, _ = m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(cliModel)
	if m.cursor != 2 {
		t.Errorf("cursor after second 'down' = %d, want 2", m.cursor)
	}

	// Test down at boundary (should not go beyond)
	updatedModel, _ = m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(cliModel)
	if m.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", m.cursor)
	}

	// Test up navigation
	updatedModel, _ = m.Update(specialKey(tea.KeyUp))
	m = updatedModel.(cliModel)
	if m.cursor != 1 {
		t.Errorf("cursor after 'up' = %d, want 1", m.cursor)
	}

	// Test up to beginning
	updatedModel, _ = m.Update(specialKey(tea.KeyUp))
	m = updatedModel.(cliModel)
	if m.cursor != 0 {
		t.Errorf("cursor after second 'up' = %d, want 0", m.cursor)
	}

	// Test up at boundary (should not go negative)
	updatedModel, _ = m.Update(specialKey(tea.KeyUp))
	m = updatedModel.(cliModel)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestCliModel_TestModeToggle(t *testing.T) {
	m := cliModel{
		profiles: []model.Profile{},
		cursor:   0,
		testMode: model.EphemeralCommit,
	}

	// Toggle to KeepCommit
	updatedModel, _ := m.Update(keyPress('m'))
	m = updatedModel.(cliModel)
	if m.testMode != model.KeepCommit {
		t.Errorf("testMode after toggle = %d, want KeepCommit (%d)", m.testMode, model.KeepCommit)
	}
	if m.statusColor != "yellow" {
		t.Errorf("statusColor = %q, want 'yellow'", m.statusColor)
	}

	// Toggle back to EphemeralCommit
	updatedModel, _ = m.Update(keyPress('m'))
	m = updatedModel.(cliModel)
	if m.testMode != model.EphemeralCommit {
		t.Errorf("testMode after second toggle = %d, want EphemeralCommit (%d)", m.testMode, model.EphemeralCommit)
	}
}

func TestCliModel_EmptyProfilesHandling(t *testing.T) {
	m := cliModel{
		profiles: []model.Profile{},
		cursor:   0,
		testMode: model.EphemeralCommit,
	}

	// Try to apply with no profiles
	updatedModel, _ := m.Update(specialKey(tea.KeyEnter))
	m = updatedModel.(cliModel)
	if m.statusMessage != "No profiles available" {
		t.Errorf("statusMessage = %q, want 'No profiles available'", m.statusMessage)
	}
	if m.statusColor != "yellow" {
		t.Errorf("statusColor = %q, want 'yellow'", m.statusColor)
	}

	// Try global apply with no profiles
	updatedModel, _ = m.Update(keyPress('g'))
	m = updatedModel.(cliModel)
	if m.statusMessage != "No profiles available" {
		t.Errorf("statusMessage for global = %q, want 'No profiles available'", m.statusMessage)
	}
}

func TestCliModel_EscapeClearsTestResult(t *testing.T) {
	m := cliModel{
		profiles:          []model.Profile{},
		cursor:            0,
		testMode:          model.EphemeralCommit,
		showingTestResult: true,
		testResult: &model.TestResult{
			SignatureValid: true,
			RawOutput:      "test output",
		},
	}

	// Press escape
	updatedModel, _ := m.Update(specialKey(tea.KeyEscape))
	m = updatedModel.(cliModel)

	if m.showingTestResult {
		t.Error("showingTestResult should be false after escape")
	}
	if m.testResult != nil {
		t.Error("testResult should be nil after escape")
	}
}

func TestCliModel_ViewRendering(t *testing.T) {
	m := cliModel{
		profiles: []model.Profile{
			{Name: "Test User", Email: "test@example.com", Key: "ABC123", Type: model.GPGProfile},
		},
		cursor:        0,
		testMode:      model.EphemeralCommit,
		statusMessage: "Test status",
		statusColor:   "green",
	}

	view := m.View()
	viewStr := view.Content

	// Check that view contains expected elements
	if !strings.Contains(viewStr, "Git Signing Manager") {
		t.Error("View should contain 'Git Signing Manager'")
	}
	if !strings.Contains(viewStr, "Test User") {
		t.Error("View should contain profile name 'Test User'")
	}
	if !strings.Contains(viewStr, "test@example.com") {
		t.Error("View should contain email 'test@example.com'")
	}
	if !strings.Contains(viewStr, "[GPG]") {
		t.Error("View should contain '[GPG]' type indicator")
	}
	if !strings.Contains(viewStr, "Test status") {
		t.Error("View should contain status message")
	}
}

func TestCliModel_TestResultViewRendering(t *testing.T) {
	m := cliModel{
		profiles:          []model.Profile{},
		cursor:            0,
		testMode:          model.EphemeralCommit,
		showingTestResult: true,
		testResult: &model.TestResult{
			SignatureValid: true,
			SigningMethod:  model.GPGSigning,
			RawOutput:      "Good signature from test",
			SignatureInfo: &model.SignatureInfo{
				KeyID:  "ABC123",
				Signer: "Test User <test@example.com>",
				Valid:  true,
			},
		},
	}

	view := m.View()
	viewStr := view.Content

	if !strings.Contains(viewStr, "Test Signing Result") {
		t.Error("View should contain 'Test Signing Result'")
	}
	if !strings.Contains(viewStr, "✓ Signature Valid") {
		t.Error("View should contain '✓ Signature Valid'")
	}
	if !strings.Contains(viewStr, "ABC123") {
		t.Error("View should contain key ID 'ABC123'")
	}
	if !strings.Contains(viewStr, "escape") {
		t.Error("View should contain escape instructions")
	}
}

func TestCliModel_InvalidTestResult(t *testing.T) {
	m := cliModel{
		profiles:          []model.Profile{},
		cursor:            0,
		testMode:          model.EphemeralCommit,
		showingTestResult: true,
		testResult: &model.TestResult{
			SignatureValid: false,
			SigningMethod:  model.GPGSigning,
			RawOutput:      "Bad signature",
		},
	}

	view := m.View()
	viewStr := view.Content

	if !strings.Contains(viewStr, "✗ Signature Invalid") {
		t.Error("View should contain '✗ Signature Invalid'")
	}
}

func TestCliModel_QuitCommand(t *testing.T) {
	m := cliModel{
		profiles: []model.Profile{},
		cursor:   0,
		testMode: model.EphemeralCommit,
	}

	_, cmd := m.Update(keyPress('q'))
	if cmd == nil {
		t.Error("'q' should return a quit command")
	}
}

func TestCliModel_Init(t *testing.T) {
	m := cliModel{}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

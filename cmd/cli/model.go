package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afeldman/git-signing-manager/internal/gitcfg"
	"github.com/afeldman/git-signing-manager/internal/gpg"
	"github.com/afeldman/git-signing-manager/internal/model"
)

type cliModel struct {
	profiles          []model.Profile
	cursor            int // which profile our cursor is pointing at
	testResult        *model.TestResult
	testMode          model.TestMode
	statusMessage     string
	statusColor       string // "green", "red", "yellow"
	showingTestResult bool
}

func initialModel() cliModel {
	profiles, _ := gpg.GetProfiles()
	return cliModel{
		profiles: profiles,
		cursor:   0,
		testMode: model.EphemeralCommit,
	}
}

func (m cliModel) Init() tea.Cmd {
	return nil
}

func (m cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}

		case "t":
			// Test with ephemeral (auto-reset) mode
			m.testMode = model.EphemeralCommit
			result, err := gitcfg.TestSigning(m.testMode)
			m.testResult = result
			if err != nil {
				m.statusMessage = fmt.Sprintf("Error: %v", err)
				m.statusColor = "red"
			} else {
				m.showingTestResult = true
				if result.SignatureValid {
					m.statusMessage = "✓ Test Signing Successful (Ephemeral)"
					m.statusColor = "green"
				} else {
					m.statusMessage = "✗ Signature Invalid"
					m.statusColor = "red"
				}
			}

		case "T":
			// Test with keep commit mode
			m.testMode = model.KeepCommit
			result, err := gitcfg.TestSigning(m.testMode)
			m.testResult = result
			if err != nil {
				m.statusMessage = fmt.Sprintf("Error: %v", err)
				m.statusColor = "red"
			} else {
				m.showingTestResult = true
				if result.SignatureValid {
					m.statusMessage = "✓ Test Signing Successful (Commit Kept)"
					m.statusColor = "green"
				} else {
					m.statusMessage = "✗ Signature Invalid"
					m.statusColor = "red"
				}
			}

		case "enter":
			p := m.profiles[m.cursor]
			err := gitcfg.ApplyProfile(p, false)
			if err != nil {
				m.statusMessage = fmt.Sprintf("Error applying profile: %v", err)
				m.statusColor = "red"
			} else {
				m.statusMessage = fmt.Sprintf("Profile applied: %s <%s>", p.Name, p.Email)
				m.statusColor = "green"
			}
			m.showingTestResult = false

		case "escape":
			// Clear test result view
			m.showingTestResult = false
			m.testResult = nil

		case "m":
			// Toggle test mode
			if m.testMode == model.EphemeralCommit {
				m.testMode = model.KeepCommit
				m.statusMessage = "Test Mode: Keep Commit"
			} else {
				m.testMode = model.EphemeralCommit
				m.statusMessage = "Test Mode: Ephemeral (Auto-reset)"
			}
			m.statusColor = "yellow"
			m.showingTestResult = false
		}
	}

	return m, nil
}

func (m cliModel) View() tea.View {
	if m.showingTestResult && m.testResult != nil {
		return m.renderTestResultView()
	}
	return m.renderMainView()
}

func (m cliModel) renderMainView() tea.View {
	var s strings.Builder

	s.WriteString("Git Signing Manager (CLI)\n")
	s.WriteString("═══════════════════════════════════════\n\n")

	s.WriteString("Available Profiles:\n\n")
	for i, p := range m.profiles {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s <%s>\n", cursor, p.Name, p.Email))
	}

	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Test Mode: %s\n", gitcfg.GetTestModeString(m.testMode)))
	s.WriteString("\n")

	if m.statusMessage != "" {
		s.WriteString(fmt.Sprintf("[%s] %s\n\n", colorizeStatus(m.statusColor), m.statusMessage))
	}

	s.WriteString("Commands:\n")
	s.WriteString("  ↑/↓     navigate profiles\n")
	s.WriteString("  enter   apply selected profile\n")
	s.WriteString("  t       test signing (ephemeral)\n")
	s.WriteString("  T       test signing (keep commit)\n")
	s.WriteString("  m       toggle test mode\n")
	s.WriteString("  q       quit\n")
	return tea.NewView(s.String())
}

func (m cliModel) renderTestResultView() tea.View {
	if m.testResult == nil {
		return tea.NewView("No test result available\n\nPress 'escape' to return")
	}

	var s strings.Builder

	s.WriteString("Test Signing Result\n")
	s.WriteString("═══════════════════════════════════════\n\n")

	if m.testResult.SignatureValid {
		s.WriteString("✓ Signature Valid\n")
	} else {
		s.WriteString("✗ Signature Invalid\n")
	}

	// Test mode and signing method
	s.WriteString(fmt.Sprintf("Mode: %s\n", gitcfg.GetTestModeString(m.testMode)))
	s.WriteString(fmt.Sprintf("Method: %s\n", gitcfg.GetSigningMethodString(m.testResult.SigningMethod)))

	// Signature info
	if m.testResult.SignatureInfo != nil && m.testResult.SignatureInfo.Valid {
		if m.testResult.SignatureInfo.Signer != "" {
			s.WriteString(fmt.Sprintf("Signer: %s\n", m.testResult.SignatureInfo.Signer))
		}
		if m.testResult.SignatureInfo.KeyID != "" {
			s.WriteString(fmt.Sprintf("Key ID: %s\n", m.testResult.SignatureInfo.KeyID))
		}
	}

	// Key expiry warning
	if m.testResult.KeyExpiryWarning != "" {
		s.WriteString(fmt.Sprintf("%s\n", m.testResult.KeyExpiryWarning))
	}

	// Error message
	if m.testResult.Error != nil {
		s.WriteString(fmt.Sprintf("\nError: %v\n", m.testResult.Error))
	}

	s.WriteString("\n--- Raw Output (selected lines) ---\n")
	if m.testResult.RawOutput != "" {
		lines := strings.Split(m.testResult.RawOutput, "\n")
		// Show first 10 lines of output
		for i, line := range lines {
			if i >= 10 {
				s.WriteString(fmt.Sprintf("... (%d more lines)\n", len(lines)-i))
				break
			}
			s.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	s.WriteString("\nPress 'escape' to return to profile selection\n")

	return tea.NewView(s.String())
}

func colorizeStatus(color string) string {
	switch color {
	case "green":
		return "✓"
	case "red":
		return "✗"
	case "yellow":
		return "⚠"
	default:
		return "•"
	}
}

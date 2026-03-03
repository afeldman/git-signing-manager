package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/afeldman/git-signing-manager/internal/gitcfg"
	"github.com/afeldman/git-signing-manager/internal/gpg"
	"github.com/afeldman/git-signing-manager/internal/model"
	"github.com/afeldman/git-signing-manager/internal/ssh"
)

// loadAllProfiles loads both GPG and SSH profiles
func loadAllProfiles() ([]model.Profile, error) {
	var allProfiles []model.Profile
	var lastErr error

	// Load GPG profiles
	gpgProfiles, err := gpg.GetProfiles()
	if err != nil {
		lastErr = fmt.Errorf("GPG: %w", err)
	}
	for i := range gpgProfiles {
		gpgProfiles[i].Type = model.GPGProfile
	}
	allProfiles = append(allProfiles, gpgProfiles...)

	// Load SSH profiles
	sshProfiles, err := ssh.GetProfiles()
	if err != nil {
		if lastErr != nil {
			lastErr = fmt.Errorf("%v; SSH: %w", lastErr, err)
		} else {
			lastErr = fmt.Errorf("SSH: %w", err)
		}
	}
	for i := range sshProfiles {
		sshProfiles[i].Type = model.SSHProfile
	}
	allProfiles = append(allProfiles, sshProfiles...)

	return allProfiles, lastErr
}

func main() {
	a := app.New()
	w := a.NewWindow("Git Signing Manager")

	profiles, err := loadAllProfiles()
	if err != nil {
		// Show warning but continue - we may have some profiles
		dialog.ShowInformation("Warning", fmt.Sprintf("Some profiles could not be loaded: %v", err), w)
	}

	options := []string{}
	for _, p := range profiles {
		options = append(options, fmt.Sprintf("[%s] %s <%s>", p.Type.String(), p.Name, p.Email))
	}

	// Profile selection
	selectBox := widget.NewSelect(options, nil)
	globalCheck := widget.NewCheck("Apply globally", nil)

	applyBtn := widget.NewButton("Apply Profile", func() {
		idx := selectBox.SelectedIndex()
		if idx < 0 || idx >= len(profiles) {
			dialog.ShowError(fmt.Errorf("please select a profile first"), w)
			return
		}
		if err := gitcfg.ApplyProfile(profiles[idx], globalCheck.Checked); err != nil {
			dialog.ShowError(fmt.Errorf("failed to apply profile: %w", err), w)
			return
		}
		dialog.ShowInformation("Success", fmt.Sprintf("Applied profile: %s", options[idx]), w)
	})

	// Test mode selection
	testModeOptions := []string{"Keep Commit", "Ephemeral (Auto-reset)"}
	testModeSelect := widget.NewSelect(testModeOptions, nil)
	testModeSelect.SetSelected("Ephemeral (Auto-reset)")

	// Test button
	testBtn := widget.NewButton("Test Signing", func() {
		testMode := model.EphemeralCommit
		if testModeSelect.SelectedIndex() == 0 {
			testMode = model.KeepCommit
		}

		result, err := gitcfg.TestSigning(testMode)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// Show result in a custom dialog
		showTestResultDialog(w, result, testMode)
	})

	// Build UI
	profileSection := container.NewVBox(
		widget.NewLabel("Select Profile:"),
		selectBox,
		globalCheck,
		applyBtn,
	)

	testSection := container.NewVBox(
		widget.NewLabel("Test Signing Mode:"),
		testModeSelect,
		testBtn,
	)

	content := container.NewVBox(
		widget.NewCard("Profile Management", "", profileSection),
		widget.NewCard("Test Signing", "", testSection),
	)

	scroll := container.NewScroll(content)
	w.SetContent(scroll)
	w.Resize(fyne.NewSize(500, 600))
	w.ShowAndRun()
}

// showTestResultDialog displays a structured test result dialog
func showTestResultDialog(w fyne.Window, result *model.TestResult, mode model.TestMode) {
	items := []fyne.CanvasObject{}

	// Status line
	statusText := "✓ Signature Valid"
	if !result.SignatureValid {
		statusText = "✗ Signature Invalid"
	}

	statusLabel := widget.NewLabel(statusText)
	items = append(items, widget.NewCard("Status", "", container.NewVBox(statusLabel)))

	// Test mode and signing method
	infoText := fmt.Sprintf("Mode: %s\nMethod: %s",
		gitcfg.GetTestModeString(mode),
		gitcfg.GetSigningMethodString(result.SigningMethod))
	items = append(items, widget.NewCard("Configuration", "", widget.NewLabel(infoText)))

	// Signature info
	if result.SignatureInfo != nil && result.SignatureInfo.Valid {
		signerInfo := ""
		if result.SignatureInfo.Signer != "" {
			signerInfo += fmt.Sprintf("Signer: %s\n", result.SignatureInfo.Signer)
		}
		if result.SignatureInfo.KeyID != "" {
			signerInfo += fmt.Sprintf("Key ID: %s\n", result.SignatureInfo.KeyID)
		}
		if signerInfo != "" {
			items = append(items, widget.NewCard("Signature Info", "", widget.NewLabel(signerInfo)))
		}
	}

	// Key expiry warning
	if result.KeyExpiryWarning != "" {
		items = append(items, widget.NewCard("Warning", "", widget.NewLabel(result.KeyExpiryWarning)))
	}

	// Raw output
	if result.RawOutput != "" {
		outputLabel := widget.NewLabel(result.RawOutput)
		outputLabel.Wrapping = fyne.TextWrapWord
		scrollOutput := container.NewScroll(outputLabel)
		scrollOutput.SetMinSize(fyne.NewSize(400, 200))
		items = append(items, widget.NewCard("Raw Output", "", scrollOutput))
	}

	// Error message
	if result.Error != nil {
		items = append(items, widget.NewCard("Error", "", widget.NewLabel(fmt.Sprintf("Error: %v", result.Error))))
	}

	// Create dialog content
	content := container.NewVBox(items...)
	scroll := container.NewScroll(content)

	// Create and show dialog
	dlg := dialog.NewCustom(
		fmt.Sprintf("Test Result: %s", gitcfg.GetTestModeString(mode)),
		"Close",
		scroll,
		w,
	)
	dlg.Resize(fyne.NewSize(500, 600))
	dlg.Show()
}

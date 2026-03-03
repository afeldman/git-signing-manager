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
)

func main() {
	a := app.New()
	w := a.NewWindow("Git Signing Manager")

	profiles, _ := gpg.GetProfiles()

	options := []string{}
	for _, p := range profiles {
		options = append(options, fmt.Sprintf("%s <%s>", p.Name, p.Email))
	}

	selectBox := widget.NewSelect(options, nil)
	globalCheck := widget.NewCheck("Apply globally", nil)

	applyBtn := widget.NewButton("Apply", func() {
		idx := selectBox.SelectedIndex()
		if idx >= 0 {
			gitcfg.ApplyProfile(profiles[idx], globalCheck.Checked)
		}
	})

	testBtn := widget.NewButton("Test Signing", func() {
		output, err := gitcfg.TestSigning()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Test Result", output, w)
	})

	content := container.NewVBox(
		selectBox,
		globalCheck,
		applyBtn,
		testBtn,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()
}

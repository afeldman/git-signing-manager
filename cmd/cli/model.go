package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/afeldman/git-signing-manager/internal/gitcfg"
	"github.com/afeldman/git-signing-manager/internal/gpg"
	"github.com/afeldman/git-signing-manager/internal/model"
)

type cliModel struct {
	profiles []model.Profile
	cursor   int // which to-do list item our cursor is pointing at
}

func initialModel() cliModel {
	profiles, _ := gpg.GetProfiles()
	return cliModel{
		profiles: profiles,
		cursor:   0,
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
			output, _ := gitcfg.TestSigning()
			fmt.Println(output)

		case "enter":
			p := m.profiles[m.cursor]
			gitcfg.ApplyProfile(p, false)
		}
	}

	return m, nil
}

func (m cliModel) View() tea.View {
	s := "Git Signing Manager (CLI)\n\n"

	for i, p := range m.profiles {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s <%s>\n", cursor, p.Name, p.Email)
	}

	s += "\n↑/↓ navigate • Enter apply • t test • q quit\n"
	return tea.NewView(s)
}

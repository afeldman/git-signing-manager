package gitcfg

import (
	"bytes"
	"os/exec"

	"github.com/afeldman/git-signing-manager/internal/model"
)

func ApplyProfile(p model.Profile, global bool) error {
	scope := "--local"
	if global {
		scope = "--global"
	}

	exec.Command("git", "config", scope, "user.name", p.Name).Run()
	exec.Command("git", "config", scope, "user.email", p.Email).Run()
	exec.Command("git", "config", scope, "user.signingkey", p.Key).Run()
	exec.Command("git", "config", scope, "commit.gpgsign", "true").Run()

	return nil
}

func TestSigning() (string, error) {
	// create signed empty commit
	cmd := exec.Command("git", "commit", "--allow-empty", "-S", "-m", "Test signing")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// show signature
	outCmd := exec.Command("git", "log", "--show-signature", "-1")
	var out bytes.Buffer
	outCmd.Stdout = &out
	outCmd.Stderr = &out

	err := outCmd.Run()
	return out.String(), err
}

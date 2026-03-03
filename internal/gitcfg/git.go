package gitcfg

import (
	"os/exec"

	"github.com/afeldman/git-signing-manager/internal/model"
)

// ApplyProfile applies a signing profile to the git repository
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

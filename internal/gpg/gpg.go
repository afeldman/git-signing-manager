package gpg

import (
	"os/exec"
	"strings"

	"github.com/afeldman/git-signing-manager/internal/model"
)

func GetProfiles() ([]model.Profile, error) {
	cmd := exec.Command("gpg", "--list-secret-keys", "--with-colons")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var profiles []model.Profile
	var key string

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 10 {
			continue
		}

		switch parts[0] {
		case "sec":
			key = parts[4]

		case "uid":
			uid := parts[9]
			if strings.Contains(uid, "<") {
				start := strings.Index(uid, "<")
				end := strings.Index(uid, ">")
				name := strings.TrimSpace(uid[:start])
				email := uid[start+1 : end]

				profiles = append(profiles, model.Profile{
					Name:  name,
					Email: email,
					Key:   key,
				})
			}
		}
	}

	return profiles, nil
}

package buildpackrunner

import (
	"fmt"
	"net/url"
	"os/exec"
)

func GitClone(repo url.URL, destination string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	branch := repo.Fragment
	repo.Fragment = ""
	gitUrl := repo.String()

	args := []string{
		"clone",
		"-depth",
		"1",
	}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	args = append(args, "--recursive", gitUrl, destination)
	cmd := exec.Command(gitPath, args...)

	err = cmd.Run()

	if err != nil {
		cmd = exec.Command(gitPath, "clone", "--recursive", gitUrl, destination)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Failed to clone git repository at %s", gitUrl)
		}

		if branch != "" {
			cmd = exec.Command(gitPath, "--git-dir="+destination+"/.git", "--work-tree="+destination, "checkout", branch)
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("Failed to checkout branch '%s' for git repository at %s", branch, gitUrl)
			}
		}
	}

	return nil
}

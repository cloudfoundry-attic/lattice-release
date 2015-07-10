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

	err = performGitClone(gitPath,
		[]string{
			"--depth",
			"1",
			"--recursive",
			gitUrl,
			destination,
		}, branch)

	if err != nil {
		err = performGitClone(gitPath,
			[]string{
				"--recursive",
				gitUrl,
				destination,
			}, branch)

		if err != nil {
			return fmt.Errorf("Failed to clone git repository at %s", gitUrl)
		}
	}

	return nil
}

func performGitClone(gitPath string, args []string, branch string) error {
	args = append([]string{"clone"}, args...)

	if branch != "" {
		args = append(args, "-b", branch)
	}
	cmd := exec.Command(gitPath, args...)
	return cmd.Run()
}

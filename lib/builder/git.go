package builder

import (
	"fmt"
	"os"

	"code.gitea.io/git"
	"io"
	"strings"
)

func cloneRepo(f io.Writer, uri, branch, reference, dir string) (string, error) {
	fmt.Fprintf(f, "Cloning %s, branch %s... ", uri, branch)

	err := git.Clone(uri, dir, git.CloneRepoOptions{
		Bare:   false,
		Branch: branch,
		Quiet:  true,
	})
	fmt.Fprintf(f, "done\n")

	if err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	fmt.Fprintf(f, "checkout reference: %s... ", reference)
	err = git.Checkout(dir, git.CheckoutOptions{Branch: reference})
	if err != nil {
		return "", err
	}
	fmt.Fprintf(f, "done\n")

	if _, err := os.Stat(dir + "/.gitmodules"); err == nil {
		fmt.Fprintf(f, "update submodules... ")
		cmd := git.NewCommand("submodule", "update", "--init", "--recursive")
		_, err := cmd.RunInDir(dir)
		if err != nil {
			return "", err
		}
		fmt.Fprint(f, "done\n")
	}

	tagcmd := git.NewCommand("describe", "--exact-match", "--tags")
	return tagcmd.RunInDir(dir)
}

func getLastCommitMessage(dir string) string {
	c := git.NewCommand("log", "-1", "--pretty=%B")
	s, err := c.RunInDir(dir)
	if err == nil {
		return strings.Replace(s, "\n", " ", -1)
	}
	return ""
}

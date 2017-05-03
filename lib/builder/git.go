package builder

import (
	"fmt"
	"io"
	"os"
	"strings"

	"code.gitea.io/git"
)

func cloneRepo(f io.Writer, uri, branch, reference, dir string) error {
	fmt.Fprintf(f, "Cloning %s, branch %s... ", uri, branch)

	err := git.Clone(uri, dir, git.CloneRepoOptions{
		Bare:   false,
		Branch: branch,
		Quiet:  true,
	})
	fmt.Fprintf(f, "done\n")

	if err != nil {
		return fmt.Errorf("git clone failed: %v", err)
	}

	fmt.Fprintf(f, "checkout reference: %s... ", reference)
	err = git.Checkout(dir, git.CheckoutOptions{Branch: reference})
	if err != nil {
		fmt.Fprintf(f, "failure\n")
		return err
	}
	fmt.Fprintf(f, "done\n")

	if _, err := os.Stat(dir + "/.gitmodules"); err == nil {
		fmt.Fprintf(f, "update submodules... ")
		cmd := git.NewCommand("submodule", "update", "--init", "--recursive")
		_, err := cmd.RunInDir(dir)
		if err != nil {
			fmt.Fprintf(f, "failure\n")
			return err
		}
		fmt.Fprint(f, "done\n")
	}

	return nil
}

func getTag(dir string) string {
	tagcmd := git.NewCommand("describe", "--exact-match", "--tags")
	tag, err := tagcmd.RunInDir(dir)
	if err == nil {
		return strings.TrimSpace(tag)
	}
	return ""
}

func getLastCommitMessage(dir string) string {
	c := git.NewCommand("log", "-1", "--pretty=%B")
	s, err := c.RunInDir(dir)
	if err == nil {
		return strings.Replace(s, "\n", " ", -1)
	}
	return ""
}

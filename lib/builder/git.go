package builder

import (
	"code.gitea.io/git"
	"fmt"
	"os"
)

func cloneRepo(f *os.File, uri, branch, reference, dir string) error {
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
		return err
	}
	fmt.Fprintf(f, "done\n")

	if _, err := os.Stat(dir + "/.gitmodules"); err == nil {
		fmt.Fprintf(f, "update submodules... ")
		cmd := git.NewCommand("submodule")
		cmd.AddArguments("update", "--init", "--recursive")
		ret, err := cmd.RunInDir(dir)
		fmt.Print("submodule shit: " + ret)
		if err != nil {
			return err
		}
		fmt.Fprint(f, "done\n")
	}

	return nil
}

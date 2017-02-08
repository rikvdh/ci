package builder

import (
	"fmt"
	"code.gitea.io/git"
)

func cloneRepo(uri, branch, reference, dir string) error {
	fmt.Println("Cloning", uri,"(", reference,")")

	err := git.Clone(uri, dir, git.CloneRepoOptions{
		Bare: false,
		Branch: branch,
		Quiet: true,
	})

	if err != nil {
		return err
	}

	err = git.Checkout(dir, git.CheckoutOptions{Branch: reference})

	if err != nil {
		return err
	}

	return nil
}

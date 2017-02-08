package indexer

import (
	"fmt"

	"srcd.works/go-git.v4"
	"srcd.works/go-git.v4/config"
	"srcd.works/go-git.v4/plumbing"
	"srcd.works/go-git.v4/storage/memory"
)

func LsRemote(repo string) error {
	// Create a new repository
	r, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return err
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "r",
		URL:  repo,
	})
	if err != nil {
		return err
	}

	rem, err := r.Remote("r")
	if err != nil {
		return err
	}

	err = rem.Fetch(&git.FetchOptions{})
	if err != nil {
		return err
	}

	refs, err := r.References()
	if err != nil {
		return err
	}

	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			fmt.Println(ref)
		}
		return nil
	})

	return nil

/*	// Pull using the create repository
	Info("git pull example")
	err = r.Pull(&git.PullOptions{
		RemoteName: "example",
	})

	CheckIfError(err)

	// List the branches
	// > git show-ref
	Info("git show-ref")

	refs, err := r.References()
	CheckIfError(err)

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// The HEAD is omitted in a `git show-ref` so we ignore the symbolic
		// references, the HEAD
		if ref.Type() == plumbing.SymbolicReference {
			return nil
		}

		fmt.Println(ref)
		return nil
	})

	CheckIfError(err)

	// Delete the example remote
	Info("git remote rm example")

	err = r.DeleteRemote("example")
	CheckIfError(err)*/
}

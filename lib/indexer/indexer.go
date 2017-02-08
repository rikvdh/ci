package indexer

import (
	"srcd.works/go-git.v4"
	"srcd.works/go-git.v4/config"
	"srcd.works/go-git.v4/plumbing"
	"srcd.works/go-git.v4/storage/memory"
)

type Branch struct {
	Hash string
	Name string
}

func RemoteBranches(repo string) ([]Branch, error) {
	// Create a new repository
	r, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return nil, err
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "r",
		URL:  repo,
	})
	if err != nil {
		return nil, err
	}

	rem, err := r.Remote("r")
	if err != nil {
		return nil, err
	}

	err = rem.Fetch(&git.FetchOptions{})
	if err != nil {
		return nil, err
	}

	refs, err := r.References()
	if err != nil {
		return nil, err
	}

	var branches []Branch

	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			branches = append(branches, Branch{
				Hash: ref.Hash().String(),
				Name: ref.Name().Short(),
			})
		}
		return nil
	})

	return branches, nil
}

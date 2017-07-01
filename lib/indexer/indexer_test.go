package indexer

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteBranchesEmpty(t *testing.T) {
	b, err := RemoteBranches("")
	assert.NotNil(t, err, "expected remote-branches to return an error")
	assert.Empty(t, b, "expected empty branch-list")
}

func TestRemoteBranchesNonExisting(t *testing.T) {
	b, err := RemoteBranches("git@github.com:boembats/knalboem.git")
	assert.NotNil(t, err, "expected remote-branches to return an error")
	assert.Empty(t, b, "expected empty branch-list")
}

func TestRemoteBranchesExisting(t *testing.T) {
	b, err := RemoteBranches("git@github.com:rikvdh/ci.git")
	assert.Nil(t, err, "expected remote-branches to return no error")
	assert.NotEmpty(t, b, "expected non-empty branch-list")
	masterFound := false
	for _, branch := range b {
		assert.NotEmpty(t, branch.Name, "branch-name may never be empty")
		assert.Len(t, branch.Hash, sha1.Size*2, "expected sha1-sum")
		_, err := hex.DecodeString(branch.Hash)
		assert.Nil(t, err, "expected valid hexstring")
		if branch.Name == "master" {
			masterFound = true
		}
	}
	assert.True(t, masterFound, "expected master to be present")
}

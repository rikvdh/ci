package main

import (
	"code.gitea.io/git"
	"fmt"
)

func main() {
	fmt.Println("aap")
	git.Clone("https://github.com/xor-gate/gopic.git", "/root/build", git.CloneRepoOptions{
		Timeout: 60,
		Branch:  "master"})
}

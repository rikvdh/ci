package main

import (
	"fmt"
	"code.gitea.io/git"
)

func main() {
	fmt.Println("aap")
	git.Clone("https://github.com/xor-gate/gopic.git", "/root/build", git.CloneRepoOptions{
		Timeout: 60,
		Branch: "master"})
}

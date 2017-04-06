package buildcfg

import (
	"net/url"
	"strings"
)

func loadGoConfig(remote string, c *Config) {
	u, _ := url.Parse(remote)
	importPath := u.Hostname() + strings.Replace(u.Path, ".git", "", 1)

	// Most part of this just moves everything from builddir to the correct go-import-path
	setup := []string{
		"export GOPATH=/build",
		"export PATH=$PATH:/build/bin",
		"mkdir /tmp/dat",
		"rsync -az . /tmp/dat",
		"rm -rf *",
		"mkdir -p src/" + importPath,
		"rsync -az /tmp/dat/ src/" + importPath,
		"rm -rf /tmp/dat",
		"cd src/" + importPath,
	}
	c.Setup.V = append(setup, c.Setup.V...)
	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "rsync")

	if len(c.Script.V) == 0 && len(c.Install.V) == 0 {
		c.Script.V = []string{
			"go get -t -v ./...",
			"go test -v ./...",
		}
	}
	c.DockerImage = "golang"
}

package buildcfg

import (
	"net/url"
	"strings"
)

func goImportPath(remote string) string {
	if strings.Contains(remote, ":") && strings.Contains(remote, "@") {
		rem := remote[strings.Index(remote, "@")+1:]
		return strings.Replace(strings.Replace(rem, ".git", "", 1), ":", "/", 1)
	}
	u, err := url.Parse(remote)
	if err != nil {
		return remote
	}
	return u.Hostname() + strings.Replace(u.Path, ".git", "", 1)
}

func loadGoConfig(remote string, c *Config) {
	importPath := goImportPath(remote)

	// Most part of this just moves everything from builddir to the correct go-import-path
	setup := []string{
		"export GOPATH=/build/.gopath",
		"export PATH=$PATH:/build/bin",
		"mkdir -p /build/.gopath/src/" + importPath,
		"mount -o bind /build /build/.gopath/src/" + importPath,
		"cd /build/.gopath/src/" + importPath,
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

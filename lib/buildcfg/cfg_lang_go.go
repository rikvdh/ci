package buildcfg

import (
	"github.com/Sirupsen/logrus"
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
	if len(c.GoImportPath) > 0 {
		c.GoImportPath = goImportPath(remote)
	}
	logrus.Infof("Goimport path for %s is %s", remote, c.GoImportPath)

	// Most part of this just moves everything from builddir to the correct go-import-path
	setup := []string{
		"export GOPATH=/build",
		"export PATH=$PATH:/build/bin",
		"cd /build/src/" + c.GoImportPath,
	}
	c.Setup.V = append(setup, c.Setup.V...)

	if len(c.Script.V) == 0 && len(c.Install.V) == 0 {
		c.Script.V = []string{
			"go get -t -v ./...",
			"go test -v ./...",
		}
	}
	c.DockerImage = "golang"
}

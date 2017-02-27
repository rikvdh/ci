// Package buildcfg is a Travis-CI compatible configuration reader
// It reads a configuration from .ci.yml or .travis.yml, it also
// selects a suitable docker image
package buildcfg

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const filename = ".ci.yml"
const travisconfig = ".travis.yml"

// Config is the travis-compatible structure
type Config struct {
	DockerImage string `yaml:"docker_image"`
	Language    string
	Addons      struct {
		Apt struct {
			Packages []string `yaml:",flow"`
		}
	}
	Setup         multiString
	BeforeInstall multiString `yaml:"before_install"`
	Install       multiString
	Script        multiString
}

// GetScript returns the build script for building the project
func (c *Config) GetScript(f *os.File) {
	f.WriteString(`#!/bin/sh
cd /build
umask 0000
apt-get -yq update
echo "$(date) Build started"
set -xe

`)
	if len(c.Addons.Apt.Packages) > 0 {
		f.WriteString("apt-get -yq --no-install-suggests --no-install-recommends --force-yes install ")
		f.WriteString(strings.Join(c.Addons.Apt.Packages, " "))
		f.WriteString("\n")
	}

	for _, s := range c.Setup.V {
		f.WriteString(s + "\n")
	}
	for _, s := range c.BeforeInstall.V {
		f.WriteString(s + "\n")
	}
	for _, s := range c.Install.V {
		f.WriteString(s + "\n")
	}
	for _, s := range c.Script.V {
		f.WriteString(s + "\n")
	}
	f.WriteString("\nset +x\necho \"$(date) Build ended\"\n")
	f.Close()
}

func loadCConfig(remote string, c *Config) {
	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "sudo", "build-essential", "cmake", "libssl-dev", "python-dev", "libffi-dev")
	c.DockerImage = "debian"
}

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
	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "sudo", "rsync")

	if len(c.Script.V) == 0 && len(c.Install.V) == 0 {
		c.Script.V = []string{
			"go get -t -v ./...",
			"go test -v ./...",
		}
	}
	c.DockerImage = "golang"
}

func loadLangConfig(language, remote string, c *Config) {
	switch language {
	case "go":
		loadGoConfig(remote, c)
	case "c":
		loadCConfig(remote, c)
	}
}

// Read build configuration, it need the root-directory for a repository
// and remote URL, some languages use this.
func Read(cfgDir, remote string) Config {
	c := Config{
		DockerImage: "debian",
	}

	d, err := ioutil.ReadFile(cfgDir + "/" + filename)
	if err != nil {
		d, err = ioutil.ReadFile(cfgDir + "/" + travisconfig)
	}

	if err != nil {
		fmt.Printf("Error reading config-file: %v\n", err)
		return c
	}

	if err := yaml.Unmarshal(d, &c); err != nil {
		fmt.Printf("yaml unmarshal failed: %v\n", err)
		return c
	}

	if c.Language != "" {
		loadLangConfig(c.Language, remote, &c)
	}

	return c
}

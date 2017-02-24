package buildcfg

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

const filename = ".ci.yml"
const travisconfig = ".travis.yml"

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

func (c *Config) GetScript(f *os.File) {
	f.WriteString(`#!/bin/sh
cd /build
echo "$(date) Build started"
set -xe

`)
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
	setup := []string{
		"apt-get update",
		"apt-get install -y --force-yes sudo build-essential cmake",
	}
	c.Setup.V = append(setup, c.Setup.V...)
	c.DockerImage = "debian"
}

func loadGoConfig(remote string, c *Config) {
	u, _ := url.Parse(remote)
	importPath := u.Hostname() + strings.Replace(u.Path, ".git", "", 1)

	setup := []string{
		"apt-get update",
		"apt-get install -y --force-yes rsync sudo",
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

	if len(c.Script.V) == 0 && len(c.Install.V) == 0 {
		c.Script.V = []string{
			"go get -t -v ./...",
			"go test -v ./...",
		}
	}
	c.DockerImage = "golang"
}

func loadLangConfig(language, remote string, c *Config) {
	c.Setup.V = append([]string{"umask 0000"}, c.Setup.V...)

	switch language {
	case "go":
		loadGoConfig(remote, c)
	case "c":
		loadCConfig(remote, c)
	}
}

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

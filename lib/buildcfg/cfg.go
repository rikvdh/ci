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
	DockerImage   string `yaml:"docker_image"`
	Language      string
	Setup         []string `yaml:",flow"`
	BeforeInstall []string `yaml:"before_install,flow"`
	Install       []string `yaml:",flow"`
	Script        []string `yaml:",flow"`
}

func (c *Config) GetScript(f *os.File) {
	f.WriteString(`#!/bin/sh
cd /build
echo "$(date) Build started"
set -xe

`)
	for _, s := range c.Setup {
		f.WriteString(s + "\n")
	}
	for _, s := range c.BeforeInstall {
		f.WriteString(s + "\n")
	}
	for _, s := range c.Install {
		f.WriteString(s + "\n")
	}
	for _, s := range c.Script {
		f.WriteString(s + "\n")
	}
	f.WriteString("\nset +x\necho \"$(date) Build ended\"\n")
	f.Close()
}

func loadLangConfig(language, remote string, c *Config) {
	if language == "go" {
		u, _ := url.Parse(remote)
		importPath := u.Hostname() + strings.Replace(u.Path, ".git", "", 1)

		setup := []string{
			"umask 0000",
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

		c.Setup = append(setup, c.Setup...)
		if len(c.Script) == 0 && len(c.Install) == 0 {
			c.Script = []string{
				"go get -t -v ./...",
				"go test -v ./...",
			}
		}
		c.DockerImage = "golang"
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
		return c
	}

	if c.Language != "" {
		loadLangConfig(c.Language, remote, &c)
	}

	return c
}

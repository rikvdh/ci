// Package buildcfg is a Travis-CI compatible configuration reader
// It reads a configuration from .ci.yml or .travis.yml, it also
// selects a suitable docker image
package buildcfg

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const filename = ".ci.yml"
const travisconfig = ".travis.yml"

// Config is the travis-compatible structure
type Config struct {
	DockerImage  string `yaml:"docker_image"`
	Language     string
	GoImportPath string `yaml:"go_import_path"`
	Addons       struct {
		Apt struct {
			Sources  []string `yaml:",flow"`
			Packages []string `yaml:",flow"`
		}
		Artifacts struct {
			Paths []string `yaml:",flow"`
		}
	}
	Setup         multiString
	BeforeInstall multiString `yaml:"before_install"`
	Install       multiString
	Script        multiString
}

func isValidCommand(cmd string) bool {
	return !strings.Contains(cmd, "goveralls")
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
	ppaInstalled := false
	needUpdate := false
	for _, alias := range c.Addons.Apt.Sources {
		source, key := translateRepo(alias)
		if strings.Contains(source, "ppa:") {
			if !ppaInstalled {
				ppaInstalled = true
				f.WriteString("apt-get install -yq --no-install-suggests --no-install-recommends --force-yes software-properties-common python-software-properties\n")
			}
			needUpdate = true
			f.WriteString("apt-add-repository -y " + source + "\n")
		} else if len(source) > 0 {
			f.WriteString("echo '" + source + "' > /etc/apt/sources.list.d/" + alias + ".list\n")
			needUpdate = true
		}
		if len(key) > 0 {
			//c.Setup.V = append(c.Setup.V, "")
		}
	}

	if needUpdate {
		f.WriteString("apt-get -yq update\n")
	}

	f.WriteString("apt-get -yq --no-install-suggests --no-install-recommends --force-yes install ")
	f.WriteString(strings.Join(c.Addons.Apt.Packages, " "))
	f.WriteString("\n")

	for _, s := range c.Setup.V {
		if isValidCommand(s) {
			f.WriteString(s + "\n")
		}
	}
	for _, s := range c.BeforeInstall.V {
		if isValidCommand(s) {
			f.WriteString(s + "\n")
		}
	}
	for _, s := range c.Install.V {
		if isValidCommand(s) {
			f.WriteString(s + "\n")
		}
	}
	for _, s := range c.Script.V {
		if isValidCommand(s) {
			f.WriteString(s + "\n")
		}
	}
	f.WriteString("\nset +x\necho \"$(date) Build ended\"\n")
	f.Close()
}

func loadLangConfig(language, remote string, c *Config) {
	switch language {
	case "go":
		loadGoConfig(remote, c)
	case "c":
		loadCConfig(remote, c)
	default:
		logrus.Errorf("Unsupported language: %s", language)
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

	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "sudo", "git-core")
	return c
}

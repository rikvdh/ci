// Package buildcfg is a Travis-CI compatible configuration reader
// It reads a configuration from .ci.yml or .travis.yml, it also
// selects a suitable docker image
package buildcfg

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	filename     = ".ci.yml"
	travisconfig = ".travis.yml"
)

// TODO: commit check for [ci skip] or [skip ci]

// Config is the travis-compatible structure
type Config struct {
	DockerImage  string `yaml:"docker_image"`
	Language     string
	GoImportPath string `yaml:"go_import_path"`
	//Branches     struct { // / == regex
	//	Only   multiString // precedece
	//	Except multiString // default skip gh-pages
	//}
	Addons struct {
		Apt struct {
			Packages []string `yaml:",flow"`
		}
		Artifacts struct {
			Paths []string `yaml:",flow"`
		}
		//Hosts multiString
	}
	//Git struct {
	//	Depth         uint
	//	LFSSkipSmudge bool `yaml:"lfs_skip_smudge"`
	//}
	globalBefore  []string
	BeforeInstall multiString `yaml:"before_install"`
	Install       multiString
	BeforeScript  multiString `yaml:"before_script"`
	Script        multiString
	//BeforeCache   multiString `yaml:"before_cache"`
	AfterSuccess multiString `yaml:"after_success"`
	AfterFailure multiString `yaml:"after_failure"`
	//BeforeDeploy  multiString `yaml:"before_deploy"`
	//Deploy        multiString
	//AfterDeploy   multiString `yaml:"after_deploy"`
	AfterScript multiString `yaml:"after_script"`
}

func isValidCommand(cmd string) bool {
	return !strings.Contains(cmd, "goveralls")
}

func writeFunc(f io.Writer, name string, script []string) {
	io.WriteString(f, name+"() (\nset -xe\n")
	for _, s := range script {
		if isValidCommand(s) {
			io.WriteString(f, s+"\n")
		}
	}
	io.WriteString(f, ") # end "+name+"\n\n")
}

// GetScript returns the build script for building the project
func (c *Config) GetScript(f io.Writer) {
	io.WriteString(f, `#!/bin/bash
cd /build
umask 0000
apt-get -yq update
echo "$(date) Build started"

`)

	io.WriteString(f, "apt-get -yq --no-install-suggests --no-install-recommends --force-yes install ")
	io.WriteString(f, strings.Join(c.Addons.Apt.Packages, " "))
	io.WriteString(f, "\n")

	for _, s := range c.globalBefore {
		io.WriteString(f, s+"\n")
	}

	writeFunc(f, "func_before_install", c.BeforeInstall.V)
	writeFunc(f, "func_install", c.Install.V)
	writeFunc(f, "func_before_script", c.BeforeScript.V)
	writeFunc(f, "func_script", c.Script.V)
	writeFunc(f, "func_after_success", c.AfterSuccess.V)
	writeFunc(f, "func_after_failure", c.AfterFailure.V)
	writeFunc(f, "func_after_script", c.AfterScript.V)

	io.WriteString(f, `
func_before_install
EXITCODE=$?
if [ $EXITCODE -ne 0 ]; then
	echo "$(date) Before install failed with code $EXITCODE"
	exit $EXITCODE;
fi

func_install
EXITCODE=$?
if [ $EXITCODE -ne 0 ]; then
	echo "$(date) Install failed with code $EXITCODE"
	exit $EXITCODE;
fi

func_before_script
EXITCODE=$?
if [ $EXITCODE -ne 0 ]; then
	echo "$(date) Before script failed with code $EXITCODE"
	exit $EXITCODE;
fi

func_script
EXITCODE=$?
if [ $EXITCODE -ne 0 ]; then
	echo "$(date) Script failed with code $EXITCODE"
	func_after_failure
else
	func_after_success
fi
func_after_script

echo "$(date) Build ended with code $EXITCODE"
exit $EXITCODE
set +x
`)
}

func loadLangConfig(c *Config, remote string) {
	switch c.Language {
	case "go":
		loadGoConfig(remote, c)
	case "c":
		loadCConfig(remote, c)
	default:
		logrus.Errorf("Unsupported language: %s", c.Language)
	}
}

func ReadConfig(config []byte, remote string) Config {
	c := Config{
		DockerImage: "debian",
	}

	if err := yaml.Unmarshal(config, &c); err != nil {
		fmt.Printf("yaml unmarshal failed: %v\n", err)
		return c
	}

	if c.Language != "" {
		loadLangConfig(&c, remote)
	}
	c.Addons.Apt.Packages = append(c.Addons.Apt.Packages, "sudo", "git-core")
	return c
}

// Read build configuration, it need the root-directory for a repository
// and remote URL, some languages use this.
func Read(cfgDir, remote string) Config {
	d, err := ioutil.ReadFile(cfgDir + "/" + filename)
	if err != nil {
		d, err = ioutil.ReadFile(cfgDir + "/" + travisconfig)
	}

	if err != nil {
		fmt.Printf("Error reading config-file: %v\n", err)
		return Config{DockerImage: "debian"}
	}

	return ReadConfig(d, remote)
}

package builder

import (
	"gopkg.in/yaml.v2"
)

type config struct {
	DockerImage string `yaml:"docker_image"`
	Script []string `yaml:",flow"`
}

func readCfg(cfgFile string) config {
	c := config{
		DockerImage: "debian",
	}

	d, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		yaml.Unmarshall(d, c)
	}

	return c
}

package buildcfg

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestConfigRead(t *testing.T) {
	s := `docker_image: test
script:
- boom
- foo`

	err := ioutil.WriteFile(".ci.yml", []byte(s), 0644)
	if err != nil {
		panic(err)
	}

	defer os.Remove(".ci.yml")

	cfg := Read(".", "boembats")

	if cfg.DockerImage != "test" {
		t.Error("Expected docker-image to be test")
	}
	if len(cfg.Script.V) != 2 {
		t.Error("Expected exactly 2 script lines")
	}
	if cfg.Script.V[0] != "boom" || cfg.Script.V[1] != "foo" {
		t.Error("expected boom foo")
	}
}

func TestDefault(t *testing.T) {
	cfg := Read("/non/existing/file", "bots")

	if cfg.DockerImage != "debian" {
		t.Error("Expected docker-image to be debian")
	}
	if len(cfg.Script.V) != 0 {
		t.Error("Expected no script lines")
	}
}

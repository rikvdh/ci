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

	err := ioutil.WriteFile("./tmp.cfg", []byte(s), 0644)
	if err != nil {
		panic(err)
	}

	defer os.Remove("./tmp.cfg")

	cfg := Read("./tmp.cfg")

	if cfg.DockerImage != "test" {
		t.Error("Expected docker-image to be test")
	}
	if len(cfg.Script) != 2 {
		t.Error("Expected exactly 2 script lines")
	}
	if cfg.Script[0] != "boom" || cfg.Script[1] != "foo" {
		t.Error("expected boom foo")
	}
}

func TestDefault(t *testing.T) {
	cfg := readCfg("/non/existing/file")

	if cfg.DockerImage != "debian" {
		t.Error("Expected docker-image to be debian")
	}
	if len(cfg.Script) != 0 {
		t.Error("Expected no script lines")
	}
}

package buildcfg

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigRead(t *testing.T) {
	s := `docker_image: test
script:
- boom
- foo`

	err := ioutil.WriteFile(".ci.yml", []byte(s), 0644)
	assert.Nil(t, err)
	defer os.Remove(".ci.yml")

	cfg := Read(".", "boembats")

	assert.Equal(t, cfg.DockerImage, "test", "Expected docker-image to be test")
	assert.Equal(t, len(cfg.Script.V), 2, "Expected exactly 2 script lines")
	assert.Equal(t, cfg.Script.V[0], "boom", "expected 1st script-item to be boom")
	assert.Equal(t, cfg.Script.V[1], "foo", "expected 2nd script-item to be foo")
}

func TestDefault(t *testing.T) {
	cfg := Read("/non/existing/file", "bots")
	assert.Equal(t, cfg.DockerImage, "debian", "Expected docker-image to be debian")
	assert.Equal(t, len(cfg.Script.V), 0, "Expected empty script")
}

func TestWildArtifactsAndApt(t *testing.T) {
	s := `language: go

go:
- go1.8
- tip

addons:
  artifacts:
    paths:
    - '*.deb'
  apt:
    packages:
    - libgsf-1-dev
    - libgsf-1-common

script:
- ./build/libvips.sh
- go get -t -v ./...
- go test -race -v $(go list ./... | grep -v /vendor/)
- go get -v -u github.com/alecthomas/gometalinter
- gometalinter --install --update
- gometalinter --deadline=600s --sort=path --vendor ./... || true
- go get -v -u github.com/xor-gate/debpkg/cmd/debpkg
- debpkg
`
	err := ioutil.WriteFile(".ci.yml", []byte(s), 0644)
	assert.Nil(t, err)

	defer os.Remove(".ci.yml")

	cfg := Read(".", "boembats")

	assert.Equal(t, cfg.DockerImage, "golang", "expected golang docker image")
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "boembats", cfg.GoImportPath)
	assert.Equal(t, "*.deb", cfg.Addons.Artifacts.Paths[0])
}

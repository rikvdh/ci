package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rikvdh/ci/lib/buildcfg"
	"github.com/rikvdh/ci/models"
	"strings"
)

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func handleArtifacts(f *os.File, job *models.Job, cfg *buildcfg.Config) {
	artifactDir := filepath.Join(buildDir, "artifacts", strconv.Itoa(int(job.ID)))
	if len(cfg.Addons.Artifacts.Paths) > 0 {
		os.Mkdir(artifactDir, 0755)
	}

	for _, artifactsPath := range cfg.Addons.Artifacts.Paths {
		paths, err := filepath.Glob(filepath.Join(job.BuildDir, artifactsPath))
		if err != nil {
			fmt.Fprintf(f, "Error searching for artifacts: %v\n", err)
			job.SetStatus(models.StatusFailed, "Missing artifacts")
			return
		}

		fmt.Println(filepath.Join(job.BuildDir, artifactsPath))
		if len(paths) == 0 {
			fmt.Fprintf(f, "Expected artifacts at %s, found none\n", artifactsPath)
			job.SetStatus(models.StatusFailed, "Missing artifacts at "+artifactsPath)
			return
		}

		for _, file := range paths {
			fi := strings.Replace(file, job.BuildDir+"/", "", 1)
			artifactFile := filepath.Join(artifactDir, fi)
			copyFile(file, artifactFile)
			models.Handle().Create(&models.Artifact{FilePath: fi, JobID: job.ID})
			fmt.Fprintf(f, "Archiving file: %s\n", fi)
		}
	}

	job.SetStatus(models.StatusPassed)
}

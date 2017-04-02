package builder

import (
	"os"

	"fmt"
	"github.com/rikvdh/ci/lib/buildcfg"
	"github.com/rikvdh/ci/models"
	"io"
	"path/filepath"
	"strconv"
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

	for _, artifacts := range cfg.Addons.Artifacts.Paths {
		paths, err := filepath.Glob(filepath.Join(job.BuildDir, artifacts))
		if err != nil {
			fmt.Fprintf(f, "Error searching for artifacts: %v\n", err)
			job.SetStatus(models.StatusFailed, "Missing artifacts")
			return
		}

		fmt.Println(filepath.Join(job.BuildDir, artifacts))
		if len(paths) == 0 {
			fmt.Fprintf(f, "Expected artifacts at %s, found none\n", artifacts)
			job.SetStatus(models.StatusFailed, "Missing artifacts at "+artifacts)
			return
		}

		for _, file := range paths {
			artifactFile := filepath.Join(artifactDir, file)
			copyFile(file, artifactFile)
			fmt.Fprintf(f, "Archiving file: %s\n", file)
		}
	}
	job.SetStatus(models.StatusPassed)
}

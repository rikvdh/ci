package web

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/sessions"
	"github.com/go-iris2/iris2/adaptors/sessions/sessiondb/file"
	"github.com/go-iris2/iris2/adaptors/view"

	"github.com/rikvdh/ci/internal/targzip"
	"github.com/rikvdh/ci/lib/auth"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/models"
)

func beforeRender(ctx *iris2.Context, m iris2.Map) iris2.Map {
	m["baseURI"] = config.Get().BaseURI
	user, err := ctx.Session().GetUint("userID")
	if err == nil {
		m["userID"] = user
	}
	return m
}

func getBranchAction(ctx *iris2.Context) {
	item, err := models.GetBranchByID(ctx.ParamInt("id"))
	if err != nil {
		emitError(ctx, "branch not found", err)
		return
	}
	user, _ := ctx.Session().GetUint("userID")

	if !((item.Build.Personal && item.Build.UserID == user) || !item.Build.Personal) {
		emitError(ctx, "permission on this branch denied", nil)
		return
	}

	var artifacts []models.Artifact
	if len(item.Jobs) > 0 {
		models.Handle().Where("job_id = ?", item.Jobs[0].ID).Find(&artifacts)

		for k := range item.Jobs {
			item.Jobs[k].Reference = item.Jobs[k].Reference[:7]
			item.Jobs[k].SetStatusTime()
		}
	}
	ctx.MustRender("branch.html", iris2.Map{
		"Page":      "Branch " + item.Name + " (" + item.Build.URI + ")",
		"Branch":    item,
		"Artifacts": artifacts})
}

func getJobAction(ctx *iris2.Context) {
	item, err := models.GetJobByID(ctx.ParamInt("id"))
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.Referer())
		return
	}
	item.Reference = item.Reference[:7]
	log := builder.GetLog(item)
	ctx.MustRender("job.html", iris2.Map{
		"Page":   "Job #" + strconv.Itoa(int(item.ID)) + " (" + item.Build.URI + ")",
		"Job":    item,
		"Log":    log,
		"LogLen": len(log)})
}

func homeAction(ctx *iris2.Context) {
	builds, err := models.BuildList(ctx.Session().GetUint("userID"))
	if err != nil {
		emitError(ctx, "error fetching build-list", err)
		return
	}
	ctx.MustRender("home.html", iris2.Map{"Page": "Home", "Builds": builds})
}

func getLatestArtifact(ctx *iris2.Context) {
	item, err := models.GetBranchByID(ctx.ParamInt("id"))
	if err != nil {
		emitError(ctx, "branch not found", err)
		return
	}

	if len(item.Jobs) == 0 {
		emitError(ctx, "no jobs for branch", nil)
		return
	}

	var artifacts []models.Artifact
	models.Handle().Where("job_id = ?", item.Jobs[0].ID).Find(&artifacts)
	if len(artifacts) == 0 {
		emitError(ctx, "no artifacts for latest job", nil)
		return
	}

	tgz, err := targzip.NewTempFile("")
	if err != nil {
		emitError(ctx, "failure creating tempfile", err)
		return
	}

	for _, artifact := range artifacts {
		fp := filepath.Join(config.Get().BuildDir, "artifacts", strconv.Itoa(int(artifact.JobID)), artifact.FilePath)
		tgz.AddFile(fp, artifact.FilePath)
	}
	tgz.Close()

	f, err := os.Open(tgz.Name())
	if err != nil {
		emitError(ctx, "opening tempfile failed", err)
		return
	}
	defer func() { f.Close(); os.Remove(tgz.Name()) }()

	if err := ctx.ServeContent(f, fmt.Sprintf("%d.tgz", item.ID), time.Now(), false); err != nil {
		emitError(ctx, "serving artifact failed", err)
		return
	}
}

func getArtifact(ctx *iris2.Context) {
	item := models.Artifact{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Where("id = ?", id).First(&item)

	fp := filepath.Join(config.Get().BuildDir, "artifacts", strconv.Itoa(int(item.JobID)), item.FilePath)
	if err := ctx.ServeFile(fp, false); err != nil {
		emitError(ctx, "serving artifact failed", err)
		return
	}
}

func Start() {
	http := iris2.New(iris2.Configuration{Gzip: false})
	http.Adapt(sessions.New(sessions.Config{
		Cookie:         "ci-session-id",
		Expires:        2 * time.Hour,
		SessionStorage: file.New("./tmp"),
	}))

	v := view.HTML("./templates", ".html")
	v.Layout("layout.html")
	v.Funcs(map[string]interface{}{
		"timeago": func(value interface{}) string {
			s := ""
			switch v := value.(type) {
			case time.Time:
				if v.Year() >= 2016 {
					s = humanize.Time(v)
				}
			case string:
				t, _ := time.Parse(time.RFC3339Nano, v)
				if t.Year() >= 2016 {
					t, err := time.Parse(time.RFC3339Nano, v)
					if err != nil {
						s = humanize.Time(t)
					}
				}
			default:
				s = fmt.Sprintf("Unknown type: %T", v)
			}
			return s
		},
	})
	http.Adapt(v)
	http.StaticWeb("/public", "./public")
	registerErrors(http)

	http.BeforeRender(beforeRender)
	http.Get("/login", loginAction)
	http.Post("/login", loginAction)

	http.Get("/register", registerAction)
	http.Post("/register", registerAction)

	http.Get("/logout", logoutAction)

	startWs(http)

	party := http.Party("", auth.New())
	{
		party.Get("/", homeAction)
		party.Get("/deletebuild/:id", deleteBuildAction)
		party.Get("/addbuild", addBuildAction)
		party.Post("/addbuild", addBuildAction)
		party.Get("/build/:id", getBuildAction)
		party.Get("/branch/:id", getBranchAction)
		party.Get("/job/:id", getJobAction)
		party.Get("/artifact/:id", getArtifact)
		party.Get("/latestartifact/:id", getLatestArtifact)
	}

	logrus.Infof("Listening on %s", config.Get().ListeningURI)
	http.Listen(config.Get().ListeningURI)
}

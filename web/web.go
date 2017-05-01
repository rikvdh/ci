package web

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ararog/timeago"
	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/sessions"
	"github.com/go-iris2/iris2/adaptors/sessions/sessiondb/file"
	"github.com/go-iris2/iris2/adaptors/view"
	"github.com/jinzhu/gorm"
	"github.com/rikvdh/ci/lib/auth"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/models"
)

func cleanReponame(remote string) string {
	if strings.Contains(remote, ":") && strings.Contains(remote, "@") {
		rem := remote[strings.Index(remote, "@")+1:]
		return strings.Replace(strings.Replace(rem, ".git", "", 1), ":", "/", 1)
	}
	u, err := url.Parse(remote)
	if err != nil {
		return remote
	}
	return u.Hostname() + strings.Replace(u.Path, ".git", "", 1)
}

func getBranchAction(ctx *iris2.Context) {
	item := models.Branch{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Preload("Jobs", func(db *gorm.DB) *gorm.DB {
		return db.Order("jobs.id DESC")
	}).Preload("Build").Where("id = ?", id).First(&item)

	item.Build.Uri = cleanReponame(item.Build.Uri)
	var artifacts []models.Artifact
	if len(item.Jobs) > 0 {
		models.Handle().Where("job_id = ?", item.Jobs[0].ID).Find(&artifacts)

		for k := range item.Jobs {
			item.Jobs[k].Reference = item.Jobs[k].Reference[:7]
			item.Jobs[k].SetStatusTime()
		}
	}
	ctx.MustRender("branch.html", iris2.Map{
		"Page":      "Branch " + item.Name + " (" + item.Build.Uri + ")",
		"Branch":    item,
		"Artifacts": artifacts})
}

func getJobAction(ctx *iris2.Context) {
	item, err := models.GetJobById(ctx.ParamInt("id"))
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.Referer())
		return
	}

	item.Build.Uri = cleanReponame(item.Build.Uri)
	item.Reference = item.Reference[:7]
	ctx.MustRender("job.html", iris2.Map{
		"Page": "Job #" + strconv.Itoa(int(item.ID)) + " (" + item.Build.Uri + ")",
		"Job":  item,
		"Log":  builder.GetLog(item)})
}

func homeAction(ctx *iris2.Context) {
	var builds []models.Build

	models.Handle().Order("updated_at DESC").Find(&builds)
	for k := range builds {
		builds[k].Uri = cleanReponame(builds[k].Uri)
	}
	ctx.MustRender("home.html", iris2.Map{"Page": "Home", "Builds": builds})
}

func beforeRender(ctx *iris2.Context, m iris2.Map) iris2.Map {
	m["baseUri"] = config.Get().BaseURI
	return m
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
		fmt.Printf("Serving artifact failed")
		ctx.Redirect(ctx.Referer())
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
					s, _ = timeago.TimeAgoFromNowWithTime(v)
				}
			case string:
				t, _ := time.Parse(time.RFC3339Nano, v)
				if t.Year() >= 2016 {
					s, _ = timeago.TimeAgoFromNowWithString(time.RFC3339Nano, v)
				}
			default:
				s = fmt.Sprintf("Unknown type: %T", v)
			}

			return s
		},
	})
	http.Adapt(v)
	http.StaticWeb("/public", "./public")

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
	}

	logrus.Infof("Listening on %s", config.Get().ListeningURI)
	http.Listen(config.Get().ListeningURI)
}

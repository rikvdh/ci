package web

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
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
	for k := range item.Jobs {
		item.Jobs[k].Reference = item.Jobs[k].Reference[:7]
		item.Jobs[k].SetStatusTime()
	}

	ctx.MustRender("branch.html", iris2.Map{"Page": "Branch " + item.Name + "(" + item.Build.Uri + ")", "Branch": item})
}

func getJobAction(ctx *iris2.Context) {
	item := models.Job{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Preload("Branch").Preload("Build").Where("id = ?", id).First(&item)

	item.SetStatusTime()

	log := builder.GetLog(&item)
	item.Build.Uri = cleanReponame(item.Build.Uri)
	item.Reference = item.Reference[:7]
	ctx.MustRender("job.html", iris2.Map{
		"Page": "Job #" + strconv.Itoa(int(item.ID)) + "(" + item.Build.Uri + ")",
		"Job":  item,
		"Log":  log})
}

func homeAction(ctx *iris2.Context) {
	var builds []models.Build

	models.Handle().Order("updated_at DESC").Find(&builds)
	for k := range builds {
		builds[k].Uri = cleanReponame(builds[k].Uri)
		builds[k].FetchLatestStatus()
	}
	ctx.MustRender("home.html", iris2.Map{"Page": "Home", "Builds": builds})
}

func beforeRender(ctx *iris2.Context, m iris2.Map) iris2.Map {
	m["baseUri"] = config.Get().BaseURI
	return m
}

func Start() {
	http := iris2.New(iris2.Configuration{Gzip: false})
	http.Adapt(sessions.New(sessions.Config{
		Cookie:         "ci-session-id",
		Expires:        2 * time.Hour,
		SessionStorage: file.New("./tmp"),
	}))
	http.Adapt(view.HTML("./templates", ".html"))
	http.Layout("layout.html")
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
	}

	logrus.Infof("Listening on %s", config.Get().ListeningURI)
	http.Listen(config.Get().ListeningURI)
}

package main

import (
	"strconv"
	"time"

	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/sessions"
	"github.com/go-iris2/iris2/adaptors/sessions/sessiondb/file"
	"github.com/go-iris2/iris2/adaptors/view"
	"github.com/jinzhu/gorm"
	"github.com/rikvdh/ci/lib/auth"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
)

func loginAction(ctx *iris2.Context) {
	user := models.User{}
	if ctx.Method() == iris2.MethodPost {
		ctx.ReadForm(&user)
		if len(user.Username) == 0 || len(user.PasswordPlain) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in your username and password")
		} else {
			models.Handle().Where(user).First(&user)
			if user.ID > 0 && user.ValidPassword() {
				ctx.Session().Set("authenticated", "true")
				ctx.Redirect("/")
				return
			}
			ctx.Session().SetFlash("msg", "Invalid credentials")
		}
	}
	ctx.MustRender("login.html", iris2.Map{"user": &user, "msg": ctx.Session().GetFlashString("msg")}, iris2.RenderOptions{"layout": iris2.NoLayout})
}

func registerAction(ctx *iris2.Context) {
	user := models.User{}

	if ctx.Method() == iris2.MethodPost {
		ctx.ReadForm(&user)
		if len(user.Username) == 0 || len(user.PasswordPlain) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in a username and password")
		} else if models.Handle().Where(user).First(&user); user.ID > 0 {
			ctx.Session().SetFlash("msg", "Duplicate user")
		} else if len(user.PasswordPlain) < 6 {
			ctx.Session().SetFlash("msg", "At least 6 characters are required for your password")
		} else {
			models.Handle().Create(&user)
			ctx.Session().SetFlash("msg", "Account created, please log-in now")
			ctx.Redirect("/login")
			return
		}
	}
	ctx.MustRender("register.html", iris2.Map{"user": &user, "msg": ctx.Session().GetFlashString("msg")}, iris2.RenderOptions{"layout": iris2.NoLayout})
}

func logoutAction(ctx *iris2.Context) {
	ctx.Session().Clear()
	ctx.Redirect("/")
}

func addBuildAction(ctx *iris2.Context) {
	build := models.Build{}

	if ctx.Method() == iris2.MethodPost {
		ctx.ReadForm(&build)
		if len(build.Uri) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in a repo URI")
		} else if models.Handle().Where(build).First(&build); build.ID > 0 {
			ctx.Session().SetFlash("msg", "Duplicate build")
		} else if _, err := indexer.RemoteBranches(build.Uri); err != nil {
			ctx.Session().SetFlash("msg", "This repository is unaccessible")
		} else {
			ctx.Session().SetFlash("msg", "Build '"+build.Uri+"' added")
			build.Status = models.StatusUnknown
			models.Handle().Create(&build)
			build = models.Build{}
		}
	}

	ctx.MustRender("add_build.html", iris2.Map{"Page": "Add build", "build": &build, "msg": ctx.Session().GetFlashString("msg")})
}

func deleteBuildAction(ctx *iris2.Context) {
	item := models.Build{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Where("id = ?", id).Delete(&item)
	ctx.Session().SetFlash("msg", "Deleted")
	ctx.Redirect(ctx.Referer())
}

func getBuildAction(ctx *iris2.Context) {
	item := models.Build{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Preload("Branches").Where("id = ?", id).First(&item)

	for k := range item.Branches {
		item.Branches[k].FetchLatestStatus()
	}

	ctx.MustRender("build.html", iris2.Map{"Page": "Build " + item.Uri, "Build": item})
}

func buildBranchAction(ctx *iris2.Context) {
	item := models.Branch{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Redirect(ctx.Referer())
		return
	}

	models.Handle().Preload("Jobs").Preload("Build").Where("id = ?", id).First(&item)
	indexer.ScheduleJob(item.Build.ID, item.ID, item.LastReference)
	ctx.Redirect(ctx.Referer())
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
	for k := range item.Jobs {
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

	ctx.MustRender("job.html", iris2.Map{
		"Page": "Job #" + strconv.Itoa(int(item.ID)) + "(" + item.Build.Uri + ")",
		"Job":  item,
		"Log":  log})
}

func homeAction(ctx *iris2.Context) {
	var builds []models.Build

	models.Handle().Find(&builds)
	for k := range builds {
		builds[k].FetchLatestStatus()
	}
	ctx.MustRender("home.html", iris2.Map{"Page": "Home", "Builds": builds})
}

func startWebinterface() {
	http := iris2.New(iris2.Configuration{Gzip: false})
	http.Adapt(sessions.New(sessions.Config{
		Cookie:         "ci-session-id",
		Expires:        2 * time.Hour,
		SessionStorage: file.New("./tmp"),
	}))
	http.Adapt(view.HTML("./templates", ".html"))
	http.Layout("layout.html")
	http.StaticWeb("/public", "./public")

	http.Get("/login", loginAction)
	http.Post("/login", loginAction)

	http.Get("/register", registerAction)
	http.Post("/register", registerAction)

	http.Get("/logout", logoutAction)

	party := http.Party("", auth.New())
	{
		party.Get("/", homeAction)
		party.Get("/deletebuild/:id", deleteBuildAction)
		party.Get("/addbuild", addBuildAction)
		party.Post("/addbuild", addBuildAction)
		party.Get("/build/:id", getBuildAction)
		party.Get("/buildbranch/:id", buildBranchAction)
		party.Get("/branch/:id", getBranchAction)
		party.Get("/job/:id", getJobAction)
	}

	http.Listen(config.Get().ListeningURI)
}

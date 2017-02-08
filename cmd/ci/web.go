package main

import (
	"github.com/kataras/go-sessions/sessiondb/file"
	"github.com/kataras/go-template/html"
	"github.com/kataras/iris"
	"github.com/rikvdh/ci/lib/auth"
	"github.com/rikvdh/ci/models"
	"github.com/rikvdh/ci/lib/indexer"
)

func loginAction(ctx *iris.Context) {
	user := models.User{}
	if ctx.Method() == "POST" {
		ctx.ReadForm(&user)
		if len(user.Username) == 0 || len(user.PasswordPlain) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in your username and password")
		} else {
			models.Handle().Where(user).First(&user)
			if user.ID > 0 && user.ValidPassword() {
				ctx.Session().Set("authenticated", "true")
				ctx.Redirect("/")
				return
			} else {
				ctx.Session().SetFlash("msg", "Invalid credentials")
			}
		}
	}
	ctx.MustRender("login.html", iris.Map{"user": &user, "msg": ctx.Session().GetFlashString("msg")}, iris.RenderOptions{"layout": iris.NoLayout})
}

func registerAction(ctx *iris.Context) {
	user := models.User{}

	if ctx.Method() == "POST" {
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
	ctx.MustRender("register.html", iris.Map{"user": &user, "msg": ctx.Session().GetFlashString("msg")}, iris.RenderOptions{"layout": iris.NoLayout})
}

func logoutAction(ctx *iris.Context) {
	ctx.Session().Clear()
	ctx.Redirect("/")
}

func addBuildAction(ctx *iris.Context) {
	build := models.Build{}

	if ctx.Method() == "POST" {
		ctx.ReadForm(&build)
		if len(build.Uri) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in a repo URI")
		} else if models.Handle().Where(build).First(&build); build.ID > 0 {
			ctx.Session().SetFlash("msg", "Duplicate build")
		} else if _, err := indexer.RemoteBranches(build.Uri); err != nil {
			ctx.Session().SetFlash("msg", "This repository is unaccessible")
		} else {
			ctx.Session().SetFlash("msg", "Build '"+build.Uri+"' added")
			models.Handle().Create(&build)
			build = models.Build{}
		}
	}

	ctx.MustRender("add_build.html", iris.Map{"Page":"Add build", "build": &build, "msg": ctx.Session().GetFlashString("msg")})
}

func deleteBuildAction(ctx *iris.Context) {
	item := models.Build{};
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.RequestHeader("Referer"))
		return
	}
	models.Handle().Where("id = ?", id).Delete(&item)
	ctx.Session().SetFlash("msg", "Deleted")
	ctx.Redirect(ctx.RequestHeader("Referer"))
}

func getBuildAction(ctx *iris.Context) {
	item := models.Build{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.RequestHeader("Referer"))
		return
	}
	models.Handle().Where("id = ?", id).First(&item)
	models.Handle().Model(&item).Related(&item.Branches)
	for k, _ := range item.Branches {
		item.Branches[k].FetchLatestStatus()
	}

	ctx.MustRender("build.html", iris.Map{"Page":"Build " + item.Uri, "Build":item})
}

func getBranchAction(ctx *iris.Context) {
	item := models.Branch{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Session().SetFlash("msg", "Invalid ID")
		ctx.Redirect(ctx.RequestHeader("Referer"))
		return
	}
	models.Handle().Where("id = ?", id).First(&item)
	models.Handle().Model(&item).Related(&item.Jobs)
	models.Handle().Model(&item).Related(&item.Build)

	ctx.MustRender("branch.html", iris.Map{"Page":"Branch " + item.Name + "("+item.Build.Uri+")", "Branch":item})
}

func homeAction(ctx *iris.Context) {
	var builds []models.Build

	models.Handle().Find(&builds)
	for k, _ := range builds {
		builds[k].FetchLatestStatus()
	}
	ctx.MustRender("home.html", iris.Map{"Page":"Home", "Builds":builds})
}

func startWebinterface() {
	cfg := iris.Configuration{
		IsDevelopment:   true,
		Gzip:            true,
		DisableBanner:   true,
		CheckForUpdates: false,
	}

	http := iris.New(cfg)
	http.UseSessionDB(file.New("../../tmp"))
	http.UseTemplate(html.New(html.Config{
		Layout: "layout.html",
	})).Directory("../../templates", ".html")

	//	http.Use(logger.New())
	http.StaticWeb("/public", "../../public")

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
		party.Get("/branch/:id", getBranchAction)
	}

	http.Listen(":8081")
}

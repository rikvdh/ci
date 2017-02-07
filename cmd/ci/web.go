package main

import (
	"github.com/kataras/go-sessions/sessiondb/file"
	"github.com/kataras/go-template/html"
	"github.com/kataras/iris"
	"github.com/rikvdh/ci/lib/auth"
	"github.com/rikvdh/ci/models"
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
			// TODO check valid repo
		} else {
			ctx.Session().SetFlash("msg", "Build '"+build.Uri+"' added")
			models.Handle().Create(&build)
			build = models.Build{}
		}
	}

	ctx.MustRender("add_build.html", iris.Map{"Page":"Add build", "build": &build, "msg": ctx.Session().GetFlashString("msg")})
}

func homeAction(ctx *iris.Context) {
	ctx.MustRender("home.html", iris.Map{"Page":"Home"})
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
		party.Get("/addbuild", addBuildAction)
		party.Post("/addbuild", addBuildAction)
	}

	http.Listen(":8081")
}

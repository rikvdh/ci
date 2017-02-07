package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/go-template/html"
	"github.com/kataras/go-sessions/sessiondb/file"
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
				ctx.Redirect("")
				return
			} else {
				ctx.Session().SetFlash("msg", "Invalid credentials")
			}
		}
	}
	ctx.MustRender("login.html", iris.Map{"user":&user,"msg":ctx.Session().GetFlashString("msg")}, iris.RenderOptions{"layout": iris.NoLayout})
}

func logoutAction(ctx *iris.Context) {
	ctx.Session().Clear()
	ctx.Redirect("/")
}

func homeAction(ctx *iris.Context) {
	ctx.MustRender("home.html", nil)
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

	http.Get("/logout", logoutAction)

	admin := http.Party("", auth.New())
	{
		admin.Get("/", homeAction)
	}

	http.Listen(":8081")
}

package web

import (
	"github.com/go-iris2/iris2"
	"github.com/rikvdh/ci/lib/config"
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
				redirectURI := ctx.Session().GetString("redirectUri")
				if len(redirectURI) == 0 {
					redirectURI = config.Get().BaseURI
				}
				ctx.Redirect(redirectURI)
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
			ctx.Redirect(config.Get().BaseURI + "login")
			return
		}
	}
	ctx.MustRender("register.html", iris2.Map{"user": &user, "msg": ctx.Session().GetFlashString("msg")}, iris2.RenderOptions{"layout": iris2.NoLayout})
}

func logoutAction(ctx *iris2.Context) {
	ctx.Session().Clear()
	ctx.Redirect(config.Get().BaseURI)
}

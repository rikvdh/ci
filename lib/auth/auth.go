package auth

import (
	"github.com/kataras/iris"
)

type authMiddleware struct {}

func (m authMiddleware) Serve(ctx *iris.Context) {
	authenticated := ctx.Session().GetString("authenticated")
	if authenticated == "true" {
		ctx.Next()
	} else {
		ctx.Redirect("/login")
	}
}

func New() iris.HandlerFunc {
	l := &authMiddleware{}
	return l.Serve
}

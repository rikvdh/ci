package auth

import (
	"github.com/go-iris2/iris2"
)

type authMiddleware struct{}

func (m authMiddleware) Serve(ctx *iris2.Context) {
	authenticated := ctx.Session().GetString("authenticated")
	if authenticated == "true" {
		ctx.Next()
	} else {
		ctx.Redirect("/login")
	}
}

func New() iris2.HandlerFunc {
	l := &authMiddleware{}
	return l.Serve
}

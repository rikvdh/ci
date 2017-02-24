package auth

import (
	"github.com/go-iris2/iris2"
)

type authMiddleware struct{}

func (m authMiddleware) Serve(ctx *iris2.Context) {
	if ctx.Session().GetString("authenticated") == "true" {
		ctx.Next()
	} else {
		ctx.Redirect("/login")
	}
}

// New returns the autnentication middleware for the web-framework
func New() iris2.HandlerFunc {
	l := &authMiddleware{}
	return l.Serve
}

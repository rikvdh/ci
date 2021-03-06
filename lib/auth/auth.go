package auth

import (
	"github.com/go-iris2/iris2"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/models"
)

type authMiddleware struct{}

func (m authMiddleware) Serve(ctx *iris2.Context) {
	if ctx.Session().GetString("authenticated") == "true" || models.ValidAPIKey(ctx.RequestHeader("X-Api-Key")) {
		ctx.Next()
	} else {

		ctx.Session().Set("redirectUri", ctx.Request.RequestURI)
		ctx.Redirect(config.Get().BaseURI + "login")
	}
}

// New returns the autnentication middleware for the web-framework
func New() iris2.HandlerFunc {
	l := &authMiddleware{}
	return l.Serve
}

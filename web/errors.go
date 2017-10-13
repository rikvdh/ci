package web

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/go-iris2/iris2"
)

func renderNotFound(ctx *iris2.Context) {
	ctx.MustRender("errors/404.html", iris2.Map{})
}

func emitError(ctx *iris2.Context, message string, err error, options ...map[string]interface{}) {
	logrus.Warnf("%s: %v", message, err)
	ctx.Set("internal_error", err)
	if len(options) > 0 {
		ctx.Set("render_options", options[0])
	}
	ctx.EmitError(http.StatusInternalServerError)
}

func renderInternalServerError(ctx *iris2.Context) {
	err, ok := ctx.Get("internal_error").(error)
	if ok {
		fmt.Printf("error is: %v\n", err)
	}
	ro, ok := ctx.Get("render_options").(map[string]interface{})
	if !ok {
		ro = nil
	}
	ctx.MustRender("errors/500.html", nil, ro)
}

func registerErrors(f *iris2.Framework) {
	f.OnError(http.StatusNotFound, renderNotFound)
	f.OnError(http.StatusInternalServerError, renderInternalServerError)
}

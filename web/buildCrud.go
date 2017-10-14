package web

import (
	"github.com/go-iris2/iris2"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
)

func addBuildAction(ctx *iris2.Context) {
	build := models.Build{}

	if ctx.Method() == iris2.MethodPost {
		if err := ctx.ReadForm(&build); err != nil {
			ctx.Session().SetFlash("msg", "Error reading form-data: "+err.Error())
		} else if err := build.IsValid(); err != nil {
			ctx.Session().SetFlash("msg", err.Error())
		} else if _, err := indexer.RemoteBranches(build.URI); err != nil {
			ctx.Session().SetFlash("msg", "could not read remote branches for URI")
		} else {
			ctx.Session().SetFlash("msg", "Build '"+build.URI+"' added")
			build.Status = models.StatusUnknown
			models.Handle().Create(&build)
			build = models.Build{}
		}
	}

	ctx.MustRender("add_build.html", iris2.Map{
		"Page":  "Add build",
		"build": &build,
		"msg":   ctx.Session().GetFlashString("msg")})
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
	b, err := ctx.ParamInt("id")
	if err != nil {
		emitError(ctx, "build is not found", err)
		return
	}
	u, err := ctx.Session().GetUint("userID")
	if err != nil {
		emitError(ctx, "user is not found", err)
		return
	}
	item, err := models.BuildWithBranches(b, u)
	if err != nil {
		emitError(ctx, "build is not found", err)
		return
	}

	ctx.MustRender("build.html", iris2.Map{"Page": "Build " + item.URI, "Build": item})
}

package web

import (
	"github.com/go-iris2/iris2"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
)

func addBuildAction(ctx *iris2.Context) {
	build := models.Build{}

	if ctx.Method() == iris2.MethodPost {
		ctx.ReadForm(&build)
		if len(build.Uri) == 0 {
			ctx.Session().SetFlash("msg", "Please fill in a repo URI")
		} else if models.Handle().Where(build).First(&build); build.ID > 0 {
			ctx.Session().SetFlash("msg", "Duplicate build")
		} else if _, err := indexer.RemoteBranches(build.Uri); err != nil {
			ctx.Session().SetFlash("msg", "This repository is unaccessible")
		} else {
			ctx.Session().SetFlash("msg", "Build '"+build.Uri+"' added")
			build.Status = models.StatusUnknown
			models.Handle().Create(&build)
			build = models.Build{}
		}
	}

	ctx.MustRender("add_build.html", iris2.Map{"Page": "Add build", "build": &build, "msg": ctx.Session().GetFlashString("msg")})
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
	item := models.Build{}
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.Redirect(ctx.Referer())
		return
	}
	models.Handle().Preload("Branches").Where("id = ?", id).First(&item)

	item.Uri = cleanReponame(item.Uri)

	for k := range item.Branches {
		item.Branches[k].FetchLatestStatus()
	}

	ctx.MustRender("build.html", iris2.Map{"Page": "Build " + item.Uri, "Build": item})
}

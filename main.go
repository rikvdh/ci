package main

import (
	"github.com/rikvdh/ci/models"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/lib/builder"
)

func main() {
	config.Load("ci.ini")
	models.Init()

	go indexer.Run()
	go builder.Run()
	startWebinterface()
}

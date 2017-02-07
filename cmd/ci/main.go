package main

import (
	"github.com/rikvdh/ci/models"
	"github.com/rikvdh/ci/lib/config"
)

func main() {
	config.Load("ci.ini")
	models.Init()

	startWebinterface()
}

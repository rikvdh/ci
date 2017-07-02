package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
	"github.com/rikvdh/ci/web"
)

var buildDate string
var buildVersion = "dev"

func main() {
	logrus.Infof("rikvdh/ci started version %s (built at %s)", buildVersion, buildDate)
	config.Load("ci.ini")
	if err := models.Init(); err != nil {
		logrus.Fatalf("model init failed: %v", err)
	}

	go indexer.Run()
	go builder.Run()
	web.Start()
}

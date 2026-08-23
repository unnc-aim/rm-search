package main

import (
	"context"
	"flag"

	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/job"
	"github.com/scutrobotlab/rm-search/server"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", "etc/config.yaml", "config file")
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	if common.IsProd() {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c := svc.ReadConfig(*configFile)
	svc.InitContext(c)

	// Apply index settings on every boot so a fresh Meilisearch instance
	// is usable immediately (filterable/sortable attributes); without
	// this, filtered searches silently degrade until setup-index runs.
	if err := index.NewIndexer(svc.Ctx()).UpdateIndexSettings(context.Background()); err != nil {
		logrus.Errorf("update index settings failed (search filters may not work): %v", err)
	}

	job.NewBase(svc.Ctx()).Start()
	server.Run(*addr)
}

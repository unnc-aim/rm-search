package main

import (
	"flag"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/job"
	"github.com/scutrobotlab/rm-search/svc"
)

func main() {
	configFile := flag.String("config", "etc/config.yaml", "config file")
	flag.Parse()

	c := svc.ReadConfig(*configFile)
	svc.InitContext(c)

	idx := index.NewIndexer(svc.Ctx())
	b := job.NewBase(svc.Ctx())
	j := job.IncrementalJob{Indexer: idx, Base: *b}
	j.Run()
}

package main

import (
	"context"
	"flag"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
)

func main() {
	configFile := flag.String("config", "etc/config.yaml", "config file")
	flag.Parse()

	c := svc.ReadConfig(*configFile)
	svc.InitContext(c)

	idx := index.NewIndexer(svc.Ctx())
	if err := idx.RecreateIndex(context.Background()); err != nil {
		panic(err)
	}
}

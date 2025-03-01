package main

import (
	"github.com/scutrobotlab/rm-search/route"
	"github.com/scutrobotlab/rm-search/svc"
)

func main() {
	c := svc.ReadConfig("etc/config.yaml")
	svc.InitContext(c)

	route.Run()
}

package main

import (
	"github.com/scutrobotlab/rm-search/server"
	"github.com/scutrobotlab/rm-search/svc"
)

func main() {
	c := svc.ReadConfig("etc/config.yaml")
	svc.InitContext(c)

	server.Run()
}

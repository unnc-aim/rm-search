package main

import (
	"flag"

	"github.com/scutrobotlab/rm-search/common"
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

	job.NewBase(svc.Ctx()).Start()
	server.Run(*addr)
}

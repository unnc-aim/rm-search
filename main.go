package main

import (
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/server"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

func main() {
	if common.IsProd() {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c := svc.ReadConfig("etc/config.yaml")
	svc.InitContext(c)

	server.Run()
}

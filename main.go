package main

import (
	"github.com/scutrobotlab/rm-search/service"
)

func main() {
	c := service.ReadConfig("etc/config.yaml")
	service.InitContext(c)
}

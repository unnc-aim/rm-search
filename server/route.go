package server

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/server/handler"
)

func Run(addr string) {
	r := gin.Default()

	r.GET("/healthz", handler.Health)
	r.POST("/ms/*path", handler.MSProxy)
	r.GET("/static/*path", handler.Static)

	g := r.Group("/statistics")
	{
		g.GET("/word-cloud", handler.WordCloud)
	}

	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}

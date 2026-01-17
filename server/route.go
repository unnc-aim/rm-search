package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/server/handler"
)

func Run(addr string) {
	r := gin.Default()

	r.GET("/healthz", handler.Health)
	r.POST("/multi-search", handler.MSearch)
	r.GET("/static/*path", handler.Static)

	g := r.Group("/statistics")
	{
		g.GET("/word-cloud", handler.WordCloud)
	}

	r.StaticFS("/", http.FS(handler.Frontend))

	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}

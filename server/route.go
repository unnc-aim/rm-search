package server

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/server/handler"
)

func Run(addr string) {
	r := gin.Default()

	r.GET("/healthz", handler.Health)
	r.POST("/ms/*path", handler.MSProxy)
	// Alias matching the production site's front-nginx prefix so clients
	// work against either deployment without a proxy layer.
	r.POST("/api/ms/*path", handler.MSProxy)
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

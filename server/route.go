package server

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/server/handler"
	"github.com/scutrobotlab/rm-search/server/handler/admin"
	"github.com/scutrobotlab/rm-search/server/middleware"
)

func Run() {
	r := gin.Default()

	r.GET("/ping", handler.Ping)
	r.POST("/_msearch", handler.MSearch)
	r.GET("/static/*path", handler.Static)

	g := r.Group("/admin", middleware.AdminAuthMiddleware())
	{
		g.GET("/ping", admin.Ping)
		g.POST("/recreate-index", admin.RecreateIndex)
	}

	err := r.Run()
	if err != nil {
		panic(err)
	}
}

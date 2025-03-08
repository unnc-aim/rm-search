package route

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/middleware"
)

func Run() {
	r := gin.Default()

	r.GET("/ping", Ping)
	r.POST("/_msearch", MSearch)
	r.GET("/static/*path", Static)

	g := r.Group("/admin", middleware.AdminAuthMiddleware())
	{
		g.GET("/ping", Ping)
	}

	err := r.Run()
	if err != nil {
		panic(err)
	}
}

package route

import (
	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()

	r.GET("/ping", Ping)
	r.POST("/_msearch", MSearch)

	err := r.Run()
	if err != nil {
		panic(err)
	}
}

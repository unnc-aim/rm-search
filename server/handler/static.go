package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/common"
	"net/http"
)

func Static(c *gin.Context) {
	path := c.Param("path")
	c.Redirect(http.StatusMovedPermanently, common.GetStaticSource(path))
}

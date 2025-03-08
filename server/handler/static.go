package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/common"
	"net/http"
	"strings"
)

func Static(c *gin.Context) {
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")
	source := common.GetStaticSource(path)
	c.Redirect(http.StatusMovedPermanently, source)
}

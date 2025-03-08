package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/common"
	"io"
	"net/http"
	"strings"
)

func Static(c *gin.Context) {
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")
	source := common.GetStaticSource(path)

	resp, err := http.Get(source)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, resp.Header.Get("Content-Type"), data)
}

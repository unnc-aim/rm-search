package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/svc"
)

func MSProxy(c *gin.Context) {
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	u, err := url.Parse(svc.Ctx().Config.MeiliSearch.Address)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	u.Path = path

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, u.String(), c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	req.Header = c.Request.Header
	req.Header.Set("Authorization", "Bearer "+svc.Ctx().Config.MeiliSearch.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.WriteHeader(resp.StatusCode)

	c.Abort()
}

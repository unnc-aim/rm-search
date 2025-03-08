package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
	"net/http"
)

func RecreateIndex(c *gin.Context) {
	ctx := c.Request.Context()
	idx := index.NewIndexer(svc.Ctx())

	err := idx.RecreateIndex(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

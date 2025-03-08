package admin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func RecreateIndex(c *gin.Context) {
	ctx := c.Request.Context()
	idx := index.NewIndexer(svc.Ctx())

	async := c.Query("async")
	asyncBool, _ := strconv.ParseBool(async)

	if asyncBool {
		go func() {
			ctx = context.Background()
			err := idx.RecreateIndex(ctx)
			if err != nil {
				logrus.Errorf("recreate index error: %v", err)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"message": "success"})
		return
	} else {
		err := idx.RecreateIndex(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success"})
		return
	}
}

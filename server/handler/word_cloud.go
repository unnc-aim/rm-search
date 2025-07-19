package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

func WordCloud(c *gin.Context) {
	var rows []common.WordCloudItem
	if value, ok := svc.Ctx().Cache.Get(common.CacheWordCloud); ok {
		rows = value.([]common.WordCloudItem)
	} else {
		logrus.Error("Cache miss for word cloud data")
		c.JSON(500, gin.H{
			"error": "Internal server error",
			"msg":   "Word cloud data not available",
		})
		return
	}

	c.JSON(200, gin.H{
		"data": rows,
	})
}

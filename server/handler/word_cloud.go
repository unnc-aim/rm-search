package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

type WordCloudItem struct {
	Word  string `json:"word" gorm:"column:query"`
	Count int64  `json:"count"`
}

func WordCloud(c *gin.Context) {
	db := svc.Ctx().Db
	query := fmt.Sprintf("SELECT `query`, COUNT(*) AS `count` FROM `search_log` GROUP BY `query` ORDER BY `count` DESC LIMIT 100")

	var rows []WordCloudItem
	if err := db.Raw(query).Scan(&rows).Error; err != nil {
		logrus.Errorf("query search log error: %v", err)
		c.JSON(500, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(200, gin.H{
		"data": rows,
	})
}

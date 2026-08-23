package job

import (
	"context"
	"github.com/patrickmn/go-cache"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

// WordCloudJob 词云生成任务
type WordCloudJob struct {
	Base
}

func (j WordCloudJob) Run() {
	ctx := context.Background()
	db := svc.Ctx().Db

	// 从数据库查询搜索日志
	const Limit = 100
	var rows []common.WordCloudItem
	query := `SELECT "query", COUNT(*) AS count FROM search_log GROUP BY "query" ORDER BY count DESC LIMIT ?`
	if err := db.WithContext(ctx).Raw(query, Limit).Scan(&rows).Error; err != nil {
		logrus.Errorf("Query search log error: %v", err)
		return
	}
	logrus.Infof("Query search log count: %d", len(rows))

	// 将结果存入缓存
	svc.Ctx().Cache.Set(common.CacheWordCloud, rows, cache.NoExpiration)
}

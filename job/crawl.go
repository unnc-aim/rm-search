package job

import (
	"context"
	"os"
	"strconv"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/sirupsen/logrus"
)

// CrawlJob 补齐爬取: 每天 0/6/12/18 点, 先把最新帖子倒序爬到追新水位,
// 再继续历史回填; 回填触底后只保留追新。详见 index/crawl.go。
type CrawlJob struct {
	Base
	Indexer *index.Indexer
}

// crawlChunkEnv bounds the ids probed per phase per run.
const crawlChunkEnv = "RM_SEARCH_CRAWL_CHUNK"

func (j CrawlJob) Run() {
	ctx := context.Background()
	chunk := int64(100_000)
	if v := os.Getenv(crawlChunkEnv); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			chunk = n
		} else {
			logrus.Warnf("invalid %s=%q, using default %d", crawlChunkEnv, v, chunk)
		}
	}

	if err := j.Indexer.ScheduledCrawl(ctx, chunk); err != nil {
		logrus.Errorf("scheduled crawl error: %v", err)
		return
	}
}

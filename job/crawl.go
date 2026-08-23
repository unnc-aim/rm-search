package job

import (
	"context"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/sirupsen/logrus"
)

// crawlChunkEnv bounds the ids probed per phase/iteration.
const crawlChunkEnv = "RM_SEARCH_CRAWL_CHUNK"

func crawlChunk() int64 {
	chunk := int64(100_000)
	if v := os.Getenv(crawlChunkEnv); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			chunk = n
		} else {
			logrus.Warnf("invalid %s=%q, using default %d", crawlChunkEnv, v, chunk)
		}
	}
	return chunk
}

// CrawlJob 追新: 每天 0/6/12/18 点 (含启动时一次), 把最新帖子倒序爬到追新水位。
// 历史回填由 BackfillLoop 常驻执行, 见下。
type CrawlJob struct {
	Base
	Indexer *index.Indexer
}

func (j CrawlJob) Run() {
	ctx := context.Background()
	if err := j.Indexer.CatchUpNewest(ctx, crawlChunk()); err != nil {
		logrus.Errorf("crawl catch-up error: %v", err)
	}
}

// BackfillLoop 历史回填: 容器启动后持续向更早的帖子推进, 每轮一个块,
// 触底后持久化 backfill_done 并永久退出; 重启后立即检测到完成标志直接退出。
type BackfillLoop struct {
	Base
	Indexer *index.Indexer
}

// Start blocks until the backfill completes or ctx is cancelled; it is
// meant to run in its own goroutine from job.Start.
func (l BackfillLoop) Start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("panic in crawl backfill loop: %v\n%s", r, debug.Stack())
		}
	}()

	for {
		done, err := l.Indexer.BackfillOnce(ctx, crawlChunk())
		if err != nil {
			logrus.Errorf("crawl backfill error: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
			}
			continue
		}
		if done {
			logrus.Info("crawl backfill already complete or just finished")
			return
		}
		// Brief pause between chunks so the loop stays lightweight and
		// reactive to shutdown.
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

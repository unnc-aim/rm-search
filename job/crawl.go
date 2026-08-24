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

// BackfillLoop 历史回填: 容器启动后以有序队列持续向更早的帖子推进
// (逐帖推进水位, 任一请求失败全体停止), 失败后冷却 5 分钟从水位线续爬;
// 触底后持久化 backfill_done 并永久退出。
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
		done, err := l.Indexer.BackfillDesc(ctx, 0)
		if err != nil {
			logrus.Errorf("crawl backfill error, cooling down 5m and resuming from watermark: %v", err)
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
		// Unbounded descent finished without done: only possible via ctx
		// cancellation; verify before looping.
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

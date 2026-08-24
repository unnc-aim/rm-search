package job

import (
	"context"
	"sync/atomic"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/sirupsen/logrus"
)

// incrementalRunning guards against overlapping runs: cron starts every
// tick in its own goroutine, and during forum outages a run can hang in
// retry loops for a long time, which used to pile up instances that then
// duplicated scans the moment the network recovered.
var incrementalRunning atomic.Bool

// IncrementalJob 增量索引
type IncrementalJob struct {
	Base
	Indexer *index.Indexer
}

func (j IncrementalJob) Run() {
	if !incrementalRunning.CompareAndSwap(false, true) {
		logrus.Warn("previous incremental run still in progress, skipping this tick")
		return
	}
	defer incrementalRunning.Store(false)

	ctx := context.Background()

	_, err := j.Indexer.IndexLatestBbsPost(ctx, "ARTICLE")
	if err != nil {
		logrus.Errorf("PersistenceLatest article error: %v", err)
		return
	}
	_, err = j.Indexer.IndexLatestBbsPost(ctx, "FAQ")
	if err != nil {
		logrus.Errorf("PersistenceLatest faq error: %v", err)
		return
	}
	_, err = j.Indexer.IndexLatestBbsPost(ctx, "WIKI")
	if err != nil {
		logrus.Errorf("PersistenceLatest wiki error: %v", err)
		return
	}
	_, err = j.Indexer.IndexLatestAnnounce(ctx)
	if err != nil {
		logrus.Errorf("PersistenceLatest announce error: %v", err)
		return
	}
}

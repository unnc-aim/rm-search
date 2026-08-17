package job

import (
	"context"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/sirupsen/logrus"
)

// IncrementalJob 增量索引
type IncrementalJob struct {
	Base
	Indexer *index.Indexer
}

func (j IncrementalJob) Run() {
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

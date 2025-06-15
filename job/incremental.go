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

	err := j.Indexer.PersistenceLatest(ctx, "ARTICLE")
	if err != nil {
		logrus.Fatalf("PersistenceLatest article error: %v", err)
	}
	err = j.Indexer.PersistenceLatest(ctx, "FAQ")
	if err != nil {
		logrus.Fatalf("PersistenceLatest faq error: %v", err)
	}
	err = j.Indexer.PersistenceLatest(ctx, "WIKI")
	if err != nil {
		logrus.Fatalf("PersistenceLatest wiki error: %v", err)
	}
}

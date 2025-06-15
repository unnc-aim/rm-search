package job

import (
	"github.com/robfig/cron/v3"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

type Base struct {
	SvcCtx *svc.Context
}

func NewBase(svcCtx *svc.Context) *Base {
	return &Base{
		SvcCtx: svcCtx,
	}
}

func (b Base) Start() {
	c := cron.New()
	indexer := index.NewIndexer(b.SvcCtx)

	_, err := c.AddJob("@every 1m", IncrementalJob{
		Base:    b,
		Indexer: indexer,
	})
	if err != nil {
		panic(err)
	}

	logrus.Infof("Init %d cron jobs", len(c.Entries()))
	c.Start()
}

package job

import (
	"github.com/robfig/cron/v3"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

type Item struct {
	Name         string   // 任务名称
	Spec         string   // Cron 表达式
	Job          cron.Job // 任务
	OnceWhenInit bool     // 是否在初始化时立即执行一次
}

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

	jobs := []Item{
		{
			Name:         "Incremental Build",
			Spec:         "@every 1m",
			Job:          IncrementalJob{Base: b, Indexer: indexer},
			OnceWhenInit: true,
		},
		{
			Name:         "Load Word Cloud",
			Spec:         "@every 5m",
			Job:          WordCloudJob{Base: b},
			OnceWhenInit: true,
		},
	}

	for _, item := range jobs {
		if item.OnceWhenInit {
			logrus.Infof("Async run job [%s] immediately", item.Name)
			go item.Job.Run()
		}
		id, err := c.AddJob(item.Spec, item.Job)
		if err != nil {
			logrus.Fatalf("Add job [%s] error: %v", item.Name, err)
		}
		logrus.Infof("Add job [%s] with ID %d", item.Name, id)
	}

	logrus.Infof("Init %d cron jobs", len(c.Entries()))
	c.Start()
}

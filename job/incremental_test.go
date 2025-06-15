package job

import (
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
	"testing"
)

func TestIncrementalJob_Run(t *testing.T) {
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	job := IncrementalJob{
		Base:    *NewBase(svcCtx),
		Indexer: index.NewIndexer(svcCtx),
	}
	job.Run()
}

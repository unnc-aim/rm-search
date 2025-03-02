package index

import "github.com/scutrobotlab/rm-search/svc"

type Indexer struct {
	SvcCtx *svc.Context
}

func NewIndexer(svcCtx *svc.Context) *Indexer {
	return &Indexer{
		SvcCtx: svcCtx,
	}
}

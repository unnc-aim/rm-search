package indexer

import "github.com/scutrobotlab/rm-search/service"

type Indexer struct {
	SvcCtx *service.Context
}

func NewIndexer(svcCtx *service.Context) *Indexer {
	return &Indexer{
		SvcCtx: svcCtx,
	}
}

package index

import (
	"github.com/scutrobotlab/rm-search/svc"
	"net/url"
)

type Indexer struct {
	SvcCtx *svc.Context
}

func NewIndexer(svcCtx *svc.Context) *Indexer {
	if svcCtx == nil {
		panic("svcCtx cannot be nil")
	}
	if svcCtx.Config.Proxy.Enabled {
		var err error
		ProxyURL, err = url.Parse(svcCtx.Config.Proxy.URL)
		if err != nil {
			panic("invalid proxy URL: " + svcCtx.Config.Proxy.URL)
		}
	}

	return &Indexer{
		SvcCtx: svcCtx,
	}
}

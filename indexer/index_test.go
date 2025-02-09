package indexer

import (
	"context"
	"github.com/scutrobotlab/rm-search/service"
	"testing"
)

func TestIndexer_ScrollAndIndex(t *testing.T) {
	ctx := context.Background()
	svcCtx := service.NewContextForTest(service.WithDb(), service.WithElastic())
	idx := NewIndexer(svcCtx)

	count, err := idx.ScrollAndIndexBbsPost(ctx, 1, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d posts", count)
}

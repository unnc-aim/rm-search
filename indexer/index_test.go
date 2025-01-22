package indexer

import (
	"context"
	"github.com/scutrobotlab/bbs-search/service"
	"math"
	"testing"
)

func TestIndexer_ScrollAndIndex(t *testing.T) {
	ctx := context.Background()
	svcCtx := service.NewContextForTest(service.WithDb(), service.WithElastic())
	idx := NewIndexer(svcCtx)

	count, err := idx.ScrollAndIndex(ctx, 1, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d posts", count)
}

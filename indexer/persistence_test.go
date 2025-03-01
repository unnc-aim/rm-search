package indexer

import (
	"context"
	"github.com/scutrobotlab/rm-search/svc"
	"testing"
)

func TestIndexer_Persistence(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb())
	idx := NewIndexer(svcCtx)
	err := idx.Persistence(ctx, 54068)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIndexer_BatchPersistence(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb())
	idx := NewIndexer(svcCtx)
	err := idx.BatchPersistenceRangeIfNotExist(ctx, 1, 11300, 10)
	if err != nil {
		t.Fatal(err)
	}
}

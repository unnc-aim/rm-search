package indexer

import (
	"context"
	"github.com/scutrobotlab/rm-search/svc"
	"testing"
)

func TestIndexer_RecreateIndex(t *testing.T) {
	TestIndexer_ScrollAndIndex(t)
	TestIndexer_DeleteUnusedIndices(t)
}

func TestIndexer_ScrollAndIndex(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	idx := NewIndexer(svcCtx)

	index, err := idx.CreateIndex()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %s created", index)

	count, err := idx.ScrollAndIndexBbsPost(ctx, index, 1, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d posts", count)
}

func TestIndexer_DeleteUnusedIndices(t *testing.T) {
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	idx := NewIndexer(svcCtx)

	if err := idx.DeleteUnusedIndices(); err != nil {
		t.Fatal(err)
	}
	t.Log("unused indices deleted")
}

func TestIndexer_ScrollAndIndexAnnounce(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	idx := NewIndexer(svcCtx)

	index, err := idx.CreateIndex()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %s created", index)

	count, err := idx.ScrollAndIndexAnnounce(ctx, index, 1, 2000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d announces", count)
}

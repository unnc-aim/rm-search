package index

import (
	"context"
	"github.com/scutrobotlab/rm-search/svc"
	"testing"
)

func TestIndexer_RecreateIndex(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	idx := NewIndexer(svcCtx)

	err := idx.RecreateIndex(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("recreate index success")
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

	err = idx.PutAlias(index)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("put alias %s", index)
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

	err = idx.PutAlias(index)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("put alias %s", index)
}

func TestIndexer_ScrollAndIndexAttachment(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithElastic())
	idx := NewIndexer(svcCtx)

	index, err := idx.CreateIndex()
	if err != nil {
		t.Fatal(err)
	}

	count, err := idx.ScrollAndIndexAttachment(ctx, index, 1, 2000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d attachments", count)

	err = idx.PutAlias(index)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("put alias %s", index)
}

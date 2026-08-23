package index

import (
	"context"
	"testing"

	"github.com/scutrobotlab/rm-search/svc"
)

func TestIndexer_RecreateIndex(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	err := idx.RecreateIndex(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("recreate index success")
}

func TestIndexer_UpdateIndexSettings(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithMeili())
	idx := NewIndexer(svcCtx)

	if err := idx.UpdateIndexSettings(ctx); err != nil {
		t.Fatal(err)
	}
	t.Log("update index settings success")
}

func TestIndexer_ScrollAndIndexBbsPost(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	count, err := idx.ScrollAndIndexBbsPost(ctx, 1, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d posts", count)
}

func TestIndexer_ScrollAndIndexAnnounce(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	count, err := idx.ScrollAndIndexAnnounce(ctx, 1, 3000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d announces", count)
}

func TestIndexer_ScrollAndIndexAttachment(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili(), svc.WithTika())
	idx := NewIndexer(svcCtx)

	count, err := idx.ScrollAndIndexAttachment(ctx, 1, 3000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d attachments", count)
}

func TestIndexer_IndexLatestAnnounce(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	count, err := idx.IndexLatestAnnounce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("index %d latest announces", count)
}

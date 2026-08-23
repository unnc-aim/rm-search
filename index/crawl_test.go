package index

import (
	"context"
	"testing"

	"github.com/scutrobotlab/rm-search/svc"
)

// TestIndexer_CrawlPhases manually exercises the catch-up and backfill
// phases with a tiny chunk against the local dev stack
// (unittest/docker-compose.yaml).
func TestIndexer_CrawlPhases(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	// Start from a clean slate so the watermark assertions hold.
	if err := svcCtx.Db.WithContext(ctx).
		Exec(`DELETE FROM crawl_state; DELETE FROM bbs_post WHERE id > 1000000;`).Error; err != nil {
		t.Fatalf("reset tables: %v", err)
	}

	// Boot path: catch-up initializes the watermarks.
	if err := idx.CatchUpNewest(ctx, 15); err != nil {
		t.Fatal(err)
	}
	st, err := idx.LoadCrawlState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("state: newest=%d oldest=%d backfill_done=%v",
		st.NewestCrawledID, st.OldestCrawledID, st.BackfillDone)

	if st.NewestCrawledID == 0 || st.OldestCrawledID == 0 {
		t.Fatalf("watermarks not initialized: %+v", st)
	}
	if st.BackfillDone {
		t.Fatal("backfill must not be done after one tiny chunk")
	}
	// Catch-up only initializes: the backfill watermark starts exactly one
	// above the newest watermark.
	if st.OldestCrawledID != st.NewestCrawledID+1 {
		t.Fatalf("unexpected watermark gap: newest=%d oldest=%d",
			st.NewestCrawledID, st.OldestCrawledID)
	}

	// Backfill iterations keep advancing the oldest watermark downwards.
	for round := 0; round < 2; round++ {
		before, err := idx.LoadCrawlState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		done, err := idx.BackfillOnce(ctx, 15)
		if err != nil {
			t.Fatal(err)
		}
		if done {
			t.Fatal("backfill must not be done near the newest posts")
		}
		after, err := idx.LoadCrawlState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if after.OldestCrawledID >= before.OldestCrawledID {
			t.Fatalf("round %d: backfill did not advance: %d -> %d",
				round, before.OldestCrawledID, after.OldestCrawledID)
		}
		if after.NewestCrawledID != before.NewestCrawledID {
			t.Fatalf("round %d: backfill touched newest watermark: %d -> %d",
				round, before.NewestCrawledID, after.NewestCrawledID)
		}
		t.Logf("round %d: oldest %d -> %d", round, before.OldestCrawledID, after.OldestCrawledID)
	}
}

package index

import (
	"context"
	"testing"

	"github.com/scutrobotlab/rm-search/svc"
)

// TestIndexer_ScheduledCrawl manually runs one scheduled crawl round with
// a tiny chunk against the local dev stack (unittest/docker-compose.yaml).
func TestIndexer_ScheduledCrawl(t *testing.T) {
	ctx := context.Background()
	svcCtx := svc.NewContextForTest(svc.WithDb(), svc.WithMeili())
	idx := NewIndexer(svcCtx)

	// Start from a clean slate so the watermark assertions hold.
	if err := svcCtx.Db.WithContext(ctx).
		Exec(`DELETE FROM crawl_state; DELETE FROM bbs_post WHERE id > 1000000;`).Error; err != nil {
		t.Fatalf("reset tables: %v", err)
	}

	if err := idx.ScheduledCrawl(ctx, 15); err != nil {
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
	if diff := st.NewestCrawledID - st.OldestCrawledID; diff < 0 || diff > 15 {
		t.Fatalf("unexpected watermark gap: %d (newest=%d oldest=%d)",
			diff, st.NewestCrawledID, st.OldestCrawledID)
	}

	// Second round continues below the oldest watermark.
	before := st.OldestCrawledID
	if err := idx.ScheduledCrawl(ctx, 15); err != nil {
		t.Fatal(err)
	}
	st2, err := idx.LoadCrawlState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st2.OldestCrawledID >= before {
		t.Fatalf("backfill did not advance: %d -> %d", before, st2.OldestCrawledID)
	}
	t.Logf("second round: newest=%d oldest=%d backfill_done=%v",
		st2.NewestCrawledID, st2.OldestCrawledID, st2.BackfillDone)
}

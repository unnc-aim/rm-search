package index

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Scheduled newest-first/backfill crawler.
//
// State is a single row in crawl_state:
//   - newest_crawled_id: everything at or above this id has been crawled
//   - oldest_crawled_id: everything below this id has been crawled
//   - backfill_done:     oldest reached the floor, historical crawling stops
//
// Each run first walks NEW posts backwards (latest -> newest+1, capped per
// run), then continues the backfill downwards from oldest_crawled_id
// (capped per run). Once the backfill hits the floor, only the newest
// catch-up phase keeps running.

// crawlFloor is the lowest post id probed by the backfill.
const crawlFloor int64 = 1

// CrawlState are the persisted watermarks of the scheduled crawler.
type CrawlState struct {
	NewestCrawledID int64
	OldestCrawledID int64
	BackfillDone    bool
}

// LoadCrawlState reads the single crawl_state row, returning zero values
// when none exists yet.
func (i *Indexer) LoadCrawlState(ctx context.Context) (*CrawlState, error) {
	var st CrawlState
	err := i.SvcCtx.Db.WithContext(ctx).
		Raw(`SELECT newest_crawled_id, oldest_crawled_id, backfill_done FROM crawl_state WHERE id = 1`).
		Scan(&st).Error
	if err != nil {
		return nil, fmt.Errorf("load crawl state: %w", err)
	}
	return &st, nil
}

// saveCrawlState upserts the watermarks.
func (i *Indexer) saveCrawlState(ctx context.Context, st *CrawlState) error {
	err := i.SvcCtx.Db.WithContext(ctx).Exec(`
INSERT INTO crawl_state (id, newest_crawled_id, oldest_crawled_id, backfill_done)
VALUES (1, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
    newest_crawled_id = excluded.newest_crawled_id,
    oldest_crawled_id = excluded.oldest_crawled_id,
    backfill_done     = excluded.backfill_done`,
		st.NewestCrawledID, st.OldestCrawledID, st.BackfillDone).Error
	if err != nil {
		return fmt.Errorf("save crawl state: %w", err)
	}
	return nil
}

// LatestPostID returns the newest post id across the crawled categories.
// Pinned posts (Top) are skipped: the list API sorts them first regardless
// of recency.
func (i *Indexer) LatestPostID(ctx context.Context) (int64, error) {
	var latest int64
	for _, category := range []string{"ARTICLE", "FAQ", "WIKI"} {
		resp, err := i.GetBbsPostList(&BbsPostListReq{
			PageSize: 20,
			PageNo:   1,
			Filter: BbsPostListReqFilter{
				Category: category,
			},
		})
		if err != nil {
			return 0, errors.Wrapf(err, "get latest %s list", category)
		}
		if resp.Code != 0 {
			return 0, fmt.Errorf("get latest %s list failed, code: %d, message: %s", category, resp.Code, resp.Message)
		}
		for _, post := range resp.Data.List {
			if post.Top {
				continue
			}
			if post.Id > latest {
				latest = post.Id
			}
			break // first non-pinned item is the category's newest
		}
	}
	if latest == 0 {
		return 0, fmt.Errorf("no posts found in any category")
	}
	return latest, nil
}

// ScheduledCrawl runs one round of the scheduled crawler. chunk bounds the
// number of ids probed in each phase so a run always finishes.
func (i *Indexer) ScheduledCrawl(ctx context.Context, chunk int64) error {
	if chunk <= 0 {
		chunk = 100_000
	}

	latest, err := i.LatestPostID(ctx)
	if err != nil {
		return err
	}

	st, err := i.LoadCrawlState(ctx)
	if err != nil {
		return err
	}

	// First run ever: nothing was missed, start the backfill right below
	// the current newest post. Posts already in the database (e.g. from a
	// manual cmd/crawl backfill) seed the watermarks so nothing is probed
	// twice.
	if st.NewestCrawledID == 0 {
		st.NewestCrawledID = latest
		st.OldestCrawledID = latest + 1
		var edge struct{ Max, Min int64 }
		if err := i.SvcCtx.Db.WithContext(ctx).
			Raw(`SELECT COALESCE(MAX(id), 0) AS max, COALESCE(MIN(id), 0) AS min FROM bbs_post`).
			Scan(&edge).Error; err != nil {
			return fmt.Errorf("seed crawl state: %w", err)
		}
		if edge.Max > st.NewestCrawledID {
			st.NewestCrawledID = edge.Max
		}
		if edge.Min > 0 && edge.Min < st.OldestCrawledID {
			st.OldestCrawledID = edge.Min
		}
		if edge.Min > 0 && edge.Min <= crawlFloor {
			st.BackfillDone = true
		}
		logrus.Infof("scheduled crawl initialized at latest post %d (db max %d, db min %d), backfill starts at %d, done=%v",
			latest, edge.Max, edge.Min, st.OldestCrawledID, st.BackfillDone)
		if err := i.saveCrawlState(ctx, st); err != nil {
			return err
		}
	}

	// Phase 1: catch up on new posts, newest first. A gap larger than
	// chunk is closed over multiple runs.
	if latest > st.NewestCrawledID {
		low := st.NewestCrawledID + 1
		if latest-low+1 > chunk {
			low = latest - chunk + 1
		}
		logrus.Infof("scheduled crawl catch-up: [%d, %d]", low, latest)
		if err := i.crawlDesc(ctx, latest, low); err != nil {
			return err
		}
		st.NewestCrawledID = latest
		if err := i.saveCrawlState(ctx, st); err != nil {
			return err
		}
	}

	// Phase 2: historical backfill, continuing below the oldest watermark
	// until the floor is reached; afterwards it never runs again.
	if !st.BackfillDone && st.OldestCrawledID > crawlFloor {
		high := st.OldestCrawledID - 1
		low := high - chunk + 1
		if low < crawlFloor {
			low = crawlFloor
		}
		logrus.Infof("scheduled crawl backfill: [%d, %d]", low, high)
		if err := i.crawlDesc(ctx, high, low); err != nil {
			return err
		}
		st.OldestCrawledID = low - 1
		if st.OldestCrawledID < crawlFloor {
			st.OldestCrawledID = crawlFloor
		}
		if low <= crawlFloor {
			st.BackfillDone = true
			logrus.Infof("scheduled crawl backfill reached floor %d, historical crawling done", crawlFloor)
		}
		if err := i.saveCrawlState(ctx, st); err != nil {
			return err
		}
	}

	return nil
}

// crawlDesc persists ids in [low, high] descending, so the newest end of
// a range lands in the database first.
func (i *Indexer) crawlDesc(ctx context.Context, high, low int64) error {
	if high < low {
		return nil
	}
	ids := make([]int64, 0, high-low+1)
	for id := high; id >= low; id-- {
		ids = append(ids, id)
	}
	return i.BatchPersistenceIds(ctx, ids, 50)
}

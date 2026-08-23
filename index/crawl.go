package index

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Watermark crawler, split by responsibility:
//
//   - CatchUpNewest: newest-first catch-up (latest -> newest watermark+1),
//     run at boot and four times a day (0/6/12/18)
//   - BackfillOnce: one bounded chunk below the oldest watermark; the
//     backfill loop in the job package calls it continuously from boot
//     until the floor is reached, after which it never runs again
//
// State lives in the single crawl_state row; newest and oldest are
// updated with separate single-column statements so the two phases can
// run concurrently without read-modify-write races.

// crawlFloor is the lowest post id probed by the backfill.
const crawlFloor int64 = 1

// crawlGoroutines bounds the watermark crawler's request concurrency.
// Deliberately below the manual crawl tool's 50: the backfill loop runs
// for hours unattended and a lower rate keeps the forum's rate limiter
// (405) and gateway (504) out of the picture.
func crawlGoroutines() int {
	if v := os.Getenv("RM_SEARCH_CRAWL_GOROUTINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		logrus.Warnf("invalid RM_SEARCH_CRAWL_GOROUTINES=%q, using default 20", v)
	}
	return 20
}

// CrawlState are the persisted watermarks of the crawler.
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

// crawlInitMu guards the one-time watermark initialization against the
// catch-up job and the backfill loop racing at boot.
var crawlInitMu sync.Mutex

// ensureCrawlState initializes the watermarks once: the catch-up starts
// from the current newest post, the backfill right below it, and posts
// already in the database (e.g. from a manual crawl) seed both ends.
func (i *Indexer) ensureCrawlState(ctx context.Context, latest int64) (*CrawlState, error) {
	crawlInitMu.Lock()
	defer crawlInitMu.Unlock()

	st, err := i.LoadCrawlState(ctx)
	if err != nil {
		return nil, err
	}
	if st.NewestCrawledID != 0 {
		return st, nil
	}

	st.NewestCrawledID = latest
	st.OldestCrawledID = latest + 1
	var edge struct{ Max, Min int64 }
	if err := i.SvcCtx.Db.WithContext(ctx).
		Raw(`SELECT COALESCE(MAX(id), 0) AS max, COALESCE(MIN(id), 0) AS min FROM bbs_post`).
		Scan(&edge).Error; err != nil {
		return nil, fmt.Errorf("seed crawl state: %w", err)
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
	logrus.Infof("crawl watermarks initialized at latest post %d (db max %d, db min %d), backfill starts at %d, done=%v",
		latest, edge.Max, edge.Min, st.OldestCrawledID, st.BackfillDone)
	if err := i.saveCrawlState(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// saveCrawlState upserts the watermarks (initialization only).
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

// saveNewestCrawled advances the newest watermark (single column, race
// free against backfill updates).
func (i *Indexer) saveNewestCrawled(ctx context.Context, id int64) error {
	err := i.SvcCtx.Db.WithContext(ctx).
		Exec(`INSERT INTO crawl_state (id, newest_crawled_id)
VALUES (1, ?)
ON CONFLICT (id) DO UPDATE SET newest_crawled_id = excluded.newest_crawled_id`, id).Error
	if err != nil {
		return fmt.Errorf("save newest crawled: %w", err)
	}
	return nil
}

// saveOldestCrawled advances the backfill watermark (single column).
func (i *Indexer) saveOldestCrawled(ctx context.Context, id int64, done bool) error {
	err := i.SvcCtx.Db.WithContext(ctx).
		Exec(`INSERT INTO crawl_state (id, oldest_crawled_id, backfill_done)
VALUES (1, ?, ?)
ON CONFLICT (id) DO UPDATE SET
    oldest_crawled_id = excluded.oldest_crawled_id,
    backfill_done     = excluded.backfill_done`, id, done).Error
	if err != nil {
		return fmt.Errorf("save oldest crawled: %w", err)
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

// CatchUpNewest walks new posts backwards from the forum's latest down to
// the newest watermark (a gap larger than chunk is closed over multiple
// runs). Run at boot and by the four-times-daily schedule.
func (i *Indexer) CatchUpNewest(ctx context.Context, chunk int64) error {
	if chunk <= 0 {
		chunk = 100_000
	}

	latest, err := i.LatestPostID(ctx)
	if err != nil {
		return err
	}
	st, err := i.ensureCrawlState(ctx, latest)
	if err != nil {
		return err
	}

	if latest <= st.NewestCrawledID {
		return nil
	}
	low := st.NewestCrawledID + 1
	if latest-low+1 > chunk {
		low = latest - chunk + 1
	}
	logrus.Infof("crawl catch-up: [%d, %d]", low, latest)
	if err := i.crawlDesc(ctx, latest, low); err != nil {
		return err
	}
	return i.saveNewestCrawled(ctx, latest)
}

// BackfillOnce crawls one bounded chunk below the oldest watermark and
// returns whether the backfill is complete. The job package calls it in
// a continuous loop from container boot.
func (i *Indexer) BackfillOnce(ctx context.Context, chunk int64) (bool, error) {
	if chunk <= 0 {
		chunk = 100_000
	}

	st, err := i.LoadCrawlState(ctx)
	if err != nil {
		return false, err
	}
	if st.NewestCrawledID == 0 {
		// Not initialized yet (boot race with the catch-up job).
		latest, err := i.LatestPostID(ctx)
		if err != nil {
			return false, err
		}
		if st, err = i.ensureCrawlState(ctx, latest); err != nil {
			return false, err
		}
	}
	if st.BackfillDone {
		return true, nil
	}

	high := st.OldestCrawledID - 1
	if high < crawlFloor {
		if err := i.saveOldestCrawled(ctx, crawlFloor, true); err != nil {
			return false, err
		}
		return true, nil
	}
	low := high - chunk + 1
	if low < crawlFloor {
		low = crawlFloor
	}
	logrus.Infof("crawl backfill chunk: [%d, %d]", low, high)
	if err := i.crawlDesc(ctx, high, low); err != nil {
		return false, err
	}

	oldest := low - 1
	if oldest < crawlFloor {
		oldest = crawlFloor
	}
	done := low <= crawlFloor
	if done {
		logrus.Infof("crawl backfill reached floor %d, historical crawling done", crawlFloor)
	}
	return done, i.saveOldestCrawled(ctx, oldest, done)
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
	return i.BatchPersistenceIds(ctx, ids, crawlGoroutines())
}

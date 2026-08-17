package index

import (
	"context"
	_ "embed"
	"math"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/scutrobotlab/rm-search/database/query"
	"github.com/sirupsen/logrus"
)

// RecreateIndex 重建索引
func (i *Indexer) RecreateIndex(ctx context.Context) error {
	logrus.Infof("recreate index start")

	err := i.UpdateIndexSettings(ctx)
	if err != nil {
		return errors.Wrapf(err, "create index error")
	}
	logrus.Infof("update index settings success")

	count, err := i.ScrollAndIndexBbsPost(ctx, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index bbs post error")
	}
	logrus.Infof("index bbs post success, count: %d", count)

	count, err = i.ScrollAndIndexAnnounce(ctx, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index announce error")
	}
	logrus.Infof("index announce success, count: %d", count)

	count, err = i.ScrollAndIndexAttachment(ctx, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index attachment error")
	}
	logrus.Infof("index attachment success, count: %d", count)

	logrus.Infof("recreate index success")
	return nil
}

func (i *Indexer) UpdateIndexSettings(ctx context.Context) error {
	_, err := i.SvcCtx.Index.UpdateSettingsWithContext(ctx, &meilisearch.Settings{
		DisplayedAttributes:  []string{"*"},
		SearchableAttributes: []string{"title", "content"},
		FilterableAttributes: []string{"source", "college_name", "category_lvl0", "category_lvl1", "create_time"},
		SortableAttributes:   []string{"create_time"},
		RankingRules: []string{
			"words",
			"typo",
			"sort",
			"attribute",
			"create_time:desc",
			"proximity",
			"exactness",
		},
		StopWords: []string{"的", "和", "如何", "实现", "开源"},
	})
	if err != nil {
		return err
	}

	return nil
}

func indexDocs[T any](i *Indexer, ctx context.Context, docs []*T, convertFunc func(*T) (*Entity, error)) (int, error) {
	entities := lo.FilterMap(docs, func(item *T, index int) (*Entity, bool) {
		e, err := convertFunc(item)
		if err != nil {
			logrus.Errorf("convert failed, obj: %v, err: %v", item, err)
			return nil, false
		}
		return e, true
	})
	if err := i.IndexDocs(ctx, entities); err != nil {
		return len(entities), errors.Wrap(err, "index docs failed")
	}

	return len(entities), nil
}

// IndexDocs 索引文档
func (i *Indexer) IndexDocs(ctx context.Context, docs []*Entity) error {
	const BatchSize = 500

	index := i.SvcCtx.Index

	t, err := index.UpdateDocumentsInBatchesWithContext(ctx, docs, BatchSize, nil)
	if err != nil {
		return errors.Wrapf(err, "update documents in batches error")
	}
	logrus.Debugf("index docs task committed: %#v", t)
	return nil
}

// ScrollAndIndexBbsPost 滚动查询并索引帖子
func (i *Indexer) ScrollAndIndexBbsPost(ctx context.Context, startId, endId int64) (int64, error) {
	const PageSize = 1000
	successCount := int64(0)
	err := i.SvcCtx.Query.Transaction(func(tx *query.Query) error {
		p := tx.BbsPost
		offset := 0
		for {
			posts, err := p.WithContext(ctx).
				Where(
					p.ID.Gte(startId),
					p.ID.Lt(endId),
					p.Code.Eq(0),
				).
				Limit(PageSize).
				Offset(offset).
				Find()
			if err != nil && !errors.Is(err, model.ErrNotFound) {
				return errors.Wrapf(err, "find posts failed, offset: %d", offset)
			}
			if len(posts) == 0 {
				break
			}
			offset += len(posts)

			succ, err := indexDocs(i, ctx, posts, ConvertBbsPost)
			successCount += int64(succ)
			if err != nil {
				return errors.Wrapf(err, "index docs failed, offset: %d", offset)
			}

			logrus.Infof("index %d posts, next offset: %d", len(posts), offset)
		}

		return nil
	})
	if err != nil {
		return successCount, errors.Wrapf(err, "transaction failed")
	}

	return successCount, nil
}

// IndexLatestBbsPost 索引最新的帖子
func (i *Indexer) IndexLatestBbsPost(ctx context.Context, category string) (int64, error) {
	b := i.SvcCtx.Query.BbsPost
	ids, err := i.PersistenceLatest(ctx, category)
	if err != nil {
		return 0, errors.Wrapf(err, "persistence latest posts failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}

	posts, err := b.WithContext(ctx).Where(b.ID.In(ids...)).Find()
	if err != nil {
		return 0, err
	}
	successCount, err := indexDocs(i, ctx, posts, ConvertBbsPost)
	if err != nil {
		return int64(successCount), errors.Wrapf(err, "index docs failed")
	}

	logrus.Infof("index latest bbs posts success for category %s, count: %d", category, successCount)

	return int64(successCount), nil
}

// IndexLatestAnnounce 索引最新的公告
func (i *Indexer) IndexLatestAnnounce(ctx context.Context) (int64, error) {
	a := i.SvcCtx.Query.Announce
	ids, err := i.PersistenceLatestAnnounce(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "persistence latest announces failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}

	announces, err := a.WithContext(ctx).Where(a.ID.In(ids...)).Find()
	if err != nil {
		return 0, err
	}
	successCount, err := indexDocs(i, ctx, announces, ConvertAnnounce)
	if err != nil {
		return int64(successCount), errors.Wrapf(err, "index docs failed")
	}

	logrus.Infof("index latest announces success, count: %d", successCount)

	return int64(successCount), nil
}

// ScrollAndIndexAnnounce 滚动查询并索引公告
func (i *Indexer) ScrollAndIndexAnnounce(ctx context.Context, startId, endId int64) (int64, error) {
	const PageSize = 1000
	successCount := int64(0)
	err := i.SvcCtx.Query.Transaction(func(tx *query.Query) error {
		a := tx.Announce
		offset := 0
		for {
			announces, err := a.WithContext(ctx).
				Where(
					a.ID.Gte(startId),
					a.ID.Lt(endId),
				).
				Limit(PageSize).
				Offset(offset).
				Find()
			if err != nil && !errors.Is(err, model.ErrNotFound) {
				return errors.Wrapf(err, "find announces failed, offset: %d", offset)
			}
			if len(announces) == 0 {
				break
			}
			offset += len(announces)

			succ, err := indexDocs(i, ctx, announces, ConvertAnnounce)
			successCount += int64(succ)
			if err != nil {
				return errors.Wrapf(err, "index docs failed, offset: %d", offset)
			}

			logrus.Infof("index %d announces, next offset: %d", len(announces), offset)
		}

		return nil
	})
	if err != nil {
		return successCount, errors.Wrapf(err, "transaction failed")
	}

	return successCount, nil
}

// ScrollAndIndexAttachment 滚动查询并索引附件
func (i *Indexer) ScrollAndIndexAttachment(ctx context.Context, startId, endId int64) (int64, error) {
	const PageSize = 1000
	successCount := int64(0)
	err := i.SvcCtx.Query.Transaction(func(tx *query.Query) error {
		p := tx.Attachment
		offset := 0
		for {
			attachments, err := p.WithContext(ctx).
				Where(
					p.ID.Gte(startId),
					p.ID.Lt(endId),
				).
				Limit(PageSize).
				Offset(offset).
				Find()
			if err != nil && !errors.Is(err, model.ErrNotFound) {
				return errors.Wrapf(err, "find attachments failed, offset: %d", offset)
			}
			if len(attachments) == 0 {
				break
			}
			offset += len(attachments)

			succ, err := indexDocs(i, ctx, attachments, ConvertAttachment)
			successCount += int64(succ)
			if err != nil {
				return errors.Wrapf(err, "index docs failed, offset: %d", offset)
			}

			logrus.Infof("index %d attachments, next offset: %d", len(attachments), offset)
		}

		return nil
	})
	if err != nil {
		return successCount, errors.Wrapf(err, "transaction failed")
	}

	return successCount, nil
}

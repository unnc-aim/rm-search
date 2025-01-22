package indexer

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/bbs-search/common"
	"log"
	"strconv"
)

// IndexDoc 索引文档
func (i *Indexer) IndexDoc(id int64, str string) error {
	elastic := i.SvcCtx.Elastic
	docId := strconv.FormatInt(id, 10)
	resp, err := elastic.Index(common.PostInfoIndex, bytes.NewBuffer([]byte(str)), elastic.Index.WithDocumentID(docId))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (i *Indexer) ScrollAndIndex(ctx context.Context, startId, endId int64) (int64, error) {
	p := i.SvcCtx.Query.PostResp

	const PageSize = 1000
	successCount := int64(0)
	for offset := startId; offset < endId; {
		posts, err := p.WithContext(ctx).
			Where(
				p.ID.Gte(offset),
				p.ID.Lt(endId),
				p.Code.Eq(0),
			).
			Limit(PageSize).
			Find()
		if err != nil {
			return successCount, errors.Wrapf(err, "find posts failed, offset: %d", offset)
		}
		if len(posts) == 0 {
			break
		}

		for _, post := range posts {
			if err = i.IndexDoc(post.ID, post.Data); err != nil {
				log.Printf("index post failed, id: %d, err: %v", post.ID, err)
				continue
			}
			successCount++
		}

		offset = posts[len(posts)-1].ID + 1
		log.Printf("index %d posts, next offset: %d", len(posts), offset)
	}

	return successCount, nil
}

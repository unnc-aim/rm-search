package indexer

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"log"
)

// IndexDoc 索引文档
func (i *Indexer) IndexDoc(id string, doc []byte) error {
	elastic := i.SvcCtx.Elastic
	resp, err := elastic.Index(common.IndexEntityName, bytes.NewBuffer(doc), elastic.Index.WithDocumentID(id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (i *Indexer) ScrollAndIndexBbsPost(ctx context.Context, startId, endId int64) (int64, error) {
	p := i.SvcCtx.Query.BbsPost

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
			id := GetEntityId(EntityTypeBbsPost, post.ID)
			doc, err := ConvertBbsPost(id, []byte(post.Data))
			if err != nil {
				log.Printf("convert post failed, id: %d, err: %v", post.ID, err)
				continue
			}
			if err = i.IndexDoc(id, doc); err != nil {
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

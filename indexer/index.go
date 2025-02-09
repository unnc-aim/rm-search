package indexer

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/sirupsen/logrus"
	"time"
)

//go:embed mapping/rm-search.json
var mapping []byte

// CreateIndex 创建索引
func (i *Indexer) CreateIndex() (string, error) {
	elastic := i.SvcCtx.Elastic
	index := fmt.Sprintf("%s-%s", common.IndexEntityName, time.Now().Format("20060102-150405"))

	// 创建索引
	resp, err := elastic.Indices.Create(index, func(req *esapi.IndicesCreateRequest) {
		req.Body = bytes.NewReader(mapping)
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 创建别名
	resp, err = elastic.Indices.PutAlias([]string{index}, common.IndexEntityName)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return index, nil
}

// IndexDoc 索引文档
func (i *Indexer) IndexDoc(index string, id string, doc []byte) error {
	elastic := i.SvcCtx.Elastic
	resp, err := elastic.Index(index, bytes.NewBuffer(doc), elastic.Index.WithDocumentID(id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (i *Indexer) ScrollAndIndexBbsPost(ctx context.Context, index string, startId, endId int64) (int64, error) {
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
				logrus.Errorf("convert post failed, id: %d, err: %v", post.ID, err)
				continue
			}
			if err = i.IndexDoc(index, id, doc); err != nil {
				logrus.Errorf("index post failed, id: %d, err: %v", post.ID, err)
				continue
			}
			successCount++
		}

		offset = posts[len(posts)-1].ID + 1
		logrus.Infof("index %d posts, next offset: %d", len(posts), offset)
	}

	return successCount, nil
}

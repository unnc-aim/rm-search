package index

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

//go:embed mapping/rm-search.json
var mapping []byte

// RecreateIndex 重建索引
func (i *Indexer) RecreateIndex(ctx context.Context) error {
	logrus.Infof("recreate index start")

	index, err := i.CreateIndex()
	if err != nil {
		return errors.Wrapf(err, "create index error")
	}
	logrus.Infof("create index %s success", index)

	count, err := i.ScrollAndIndexBbsPost(ctx, index, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index bbs post error")
	}
	logrus.Infof("index bbs post success, count: %d", count)

	count, err = i.ScrollAndIndexAnnounce(ctx, index, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index announce error")
	}
	logrus.Infof("index announce success, count: %d", count)

	count, err = i.ScrollAndIndexAttachment(ctx, index, 1, math.MaxInt64)
	if err != nil {
		return errors.Wrapf(err, "scroll and index attachment error")
	}
	logrus.Infof("index attachment success, count: %d", count)

	err = i.UpdateAlias(index)
	if err != nil {
		return errors.Wrapf(err, "update alias error")
	}
	logrus.Infof("update alias success")

	err = i.DeleteUnusedIndices()
	if err != nil {
		return errors.Wrapf(err, "delete unused indices error")
	}
	logrus.Infof("delete unused indices success")

	logrus.Infof("recreate index success")
	return nil
}

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
	if resp.StatusCode != 200 {
		return "", errors.Errorf("create index failed, status code: %d", resp.StatusCode)
	}

	return index, nil
}

// PutAlias 创建别名
func (i *Indexer) PutAlias(index string) error {
	elastic := i.SvcCtx.Elastic

	// 创建别名
	resp, err := elastic.Indices.PutAlias([]string{index}, common.IndexEntityName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("put alias failed, status code: %d", resp.StatusCode)
	}

	return nil
}

// UpdateAlias 替换别名
func (i *Indexer) UpdateAlias(newIndex string) error {
	elastic := i.SvcCtx.Elastic

	// 更新别名
	resp, err := elastic.Indices.UpdateAliases(nil, func(req *esapi.IndicesUpdateAliasesRequest) {
		actions := []map[string]interface{}{
			{
				"remove": map[string]interface{}{
					"alias":   common.IndexEntityName,
					"indices": "_all",
				},
			},
			{
				"add": map[string]interface{}{
					"alias":   common.IndexEntityName,
					"indices": newIndex,
				},
			},
		}
		body := map[string]interface{}{
			"actions": actions,
		}
		req.Body = bytes.NewReader(common.MustMarshal(body))
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("update aliases failed, status code: %d", resp.StatusCode)
	}

	return nil
}

// GetLatestIndexName 获取最新的索引名称
func (i *Indexer) GetLatestIndexName() (string, error) {
	elastic := i.SvcCtx.Elastic

	// 获取别名 rm-search 关联的所有索引
	aliasResp, err := elastic.Indices.GetAlias(elastic.Indices.GetAlias.WithName(common.IndexEntityName))
	if err != nil {
		return "", errors.Wrap(err, "get alias failed")
	}
	defer aliasResp.Body.Close()

	if aliasResp.StatusCode != 200 {
		return "", errors.Errorf("get alias failed, status code: %d", aliasResp.StatusCode)
	}

	var aliasIndices map[string]interface{}
	if err := json.NewDecoder(aliasResp.Body).Decode(&aliasIndices); err != nil {
		return "", errors.Wrap(err, "decode alias response failed")
	}
	if len(aliasIndices) == 0 {
		return "", errors.New("no indices found for alias")
	}
	// 获取最新的索引名称
	var latestIndex string
	for indexName := range aliasIndices {
		if latestIndex == "" || indexName > latestIndex {
			latestIndex = indexName
		}
	}
	return latestIndex, nil
}

// DeleteUnusedIndices 删除未使用的索引
func (i *Indexer) DeleteUnusedIndices() error {
	elastic := i.SvcCtx.Elastic

	// 获取所有以 rm-search- 开头的索引
	resp, err := elastic.Cat.Indices(
		elastic.Cat.Indices.WithFormat("json"),
		elastic.Cat.Indices.WithIndex(fmt.Sprintf("%s-*", common.IndexEntityName)),
	)
	if err != nil {
		return errors.Wrapf(err, "get indices failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("get indices failed, status code: %d", resp.StatusCode)
	}

	var indices []map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&indices); err != nil {
		return err
	}

	// 获取别名 rm-search 关联的所有索引
	aliasResp, err := elastic.Indices.GetAlias(elastic.Indices.GetAlias.WithName(common.IndexEntityName))
	if err != nil {
		return err
	}
	defer aliasResp.Body.Close()
	if aliasResp.StatusCode != 200 {
		return errors.Errorf("get alias failed, status code: %d", aliasResp.StatusCode)
	}

	var aliasIndices map[string]interface{}
	if err := json.NewDecoder(aliasResp.Body).Decode(&aliasIndices); err != nil {
		return err
	}

	// 找出所有未绑定到别名 rm-search 的索引
	var unusedIndices []string
	for _, index := range indices {
		indexName := index["index"].(string)
		if _, exists := aliasIndices[indexName]; !exists {
			unusedIndices = append(unusedIndices, indexName)
		}
	}

	// 删除这些未使用的索引
	if len(unusedIndices) > 0 {
		deleteResp, err := elastic.Indices.Delete(unusedIndices)
		if err != nil {
			return err
		}
		defer deleteResp.Body.Close()
		if deleteResp.StatusCode != 200 {
			return errors.Errorf("delete indices failed, status code: %d", deleteResp.StatusCode)
		}
		logrus.Infof("delete %d unused indices", len(unusedIndices))
	} else {
		logrus.Infof("no unused indices")
	}

	return nil
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

// ScrollAndIndexBbsPost 滚动查询并索引帖子
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
			err = i.IndexBbsPost(index, post)
			if err != nil {
				if !errors.Is(err, ErrBbsPostCannotIndex) {
					logrus.Errorf("index post failed, id: %d, err: %v", post.ID, err)
				}
				continue
			}
			successCount++
		}

		offset = posts[len(posts)-1].ID + 1
		logrus.Infof("index %d posts, next offset: %d", len(posts), offset)
	}

	return successCount, nil
}

// IndexBbsPost 索引单个帖子
func (i *Indexer) IndexBbsPost(index string, item *model.BbsPost) error {
	id := GetEntityId(EntityTypeBbsPost, item.ID)
	doc, err := ConvertBbsPost(id, []byte(item.Data))
	if err != nil {
		return err
	}
	if err = i.IndexDoc(index, id, doc); err != nil {
		return err
	}
	return nil
}

// IndexLatestBbsPost 索引最新的帖子
func (i *Indexer) IndexLatestBbsPost(ctx context.Context, index string, category string) (int64, error) {
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
	var successCount int64
	for _, post := range posts {
		if post == nil {
			continue
		}
		err = i.IndexBbsPost(index, post)
		if err != nil {
			logrus.Errorf("index latest bbs post failed, id: %d, err: %v", post.ID, err)
			continue
		}
		successCount++
	}
	logrus.Infof("index latest bbs posts success for category %s, count: %d", category, successCount)

	return successCount, nil
}

// ScrollAndIndexAnnounce 滚动查询并索引公告
func (i *Indexer) ScrollAndIndexAnnounce(ctx context.Context, index string, startId, endId int64) (int64, error) {
	a := i.SvcCtx.Query.Announce

	const PageSize = 1000
	successCount := int64(0)
	for offset := startId; offset < endId; {
		announces, err := a.WithContext(ctx).
			Where(
				a.ID.Gte(offset),
				a.ID.Lt(endId),
			).
			Limit(PageSize).
			Find()
		if err != nil {
			return successCount, errors.Wrapf(err, "find announces failed, offset: %d", offset)
		}
		if len(announces) == 0 {
			break
		}
		for _, announce := range announces {
			if announce == nil {
				continue
			}
			if !announce.Found {
				continue
			}
			id := GetEntityId(EntityTypeAnnounce, announce.ID)
			doc, err := ConvertAnnounce(id, *announce)
			if err != nil {
				logrus.Errorf("convert announce failed, id: %d, err: %v", announce.ID, err)
				continue
			}
			if err = i.IndexDoc(index, id, doc); err != nil {
				logrus.Errorf("index announce failed, id: %d, err: %v", announce.ID, err)
				continue
			}
			successCount++
		}

		offset = announces[len(announces)-1].ID + 1
		logrus.Infof("index %d announces, next offset: %d", len(announces), offset)
	}

	return successCount, nil
}

// ScrollAndIndexAttachment 滚动查询并索引附件
func (i *Indexer) ScrollAndIndexAttachment(ctx context.Context, index string, startId, endId int64) (int64, error) {
	p := i.SvcCtx.Query.Attachment

	const PageSize = 1000
	successCount := int64(0)
	for offset := startId; offset < endId; {
		attachments, err := p.WithContext(ctx).
			Where(
				p.ID.Gte(offset),
				p.ID.Lt(endId),
			).
			Limit(PageSize).
			Find()
		if err != nil {
			return successCount, errors.Wrapf(err, "find attachments failed, offset: %d", offset)
		}
		if len(attachments) == 0 {
			break
		}
		for _, attachment := range attachments {
			id := GetEntityId(EntityTypeAttachment, attachment.ID)
			doc, err := ConvertAttachment(id, *attachment)
			if err != nil {
				logrus.Errorf("convert attachment failed, id: %d, err: %v", attachment.ID, err)
				continue
			}
			if err = i.IndexDoc(index, id, doc); err != nil {
				logrus.Errorf("index attachment failed, id: %d, err: %v", attachment.ID, err)
				continue
			}
			successCount++
		}

		offset = attachments[len(attachments)-1].ID + 1
		logrus.Infof("index %d attachments, next offset: %d", len(attachments), offset)
	}

	return successCount, nil
}

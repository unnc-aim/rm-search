package index

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/sirupsen/logrus"
	"math"
	"net/url"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

var (
	Mutex         = sync.Mutex{}
	PostCount     = atomic.Int64{}
	LastPrintTime = time.Now()
)

// BatchPersistenceRangeIfNotExist 批量持久化帖子，如果帖子不存在
func (i *Indexer) BatchPersistenceRangeIfNotExist(ctx context.Context, startId, endId int64, goroutine int) error {
	// 查询已经持久化的帖子
	p := i.SvcCtx.Query.BbsPost
	find, err := p.WithContext(ctx).
		Select(p.ID).
		Where(p.ID.Gte(startId), p.ID.Lt(endId)).
		Find()
	if err != nil {
		return errors.Wrap(err, "find post failed")
	}
	idSet := make(map[int64]struct{})
	for _, post := range find {
		idSet[post.ID] = struct{}{}
	}
	logrus.Infof("found %d posts have been persisted", len(idSet))

	// 找出未持久化的帖子
	ids := make([]int64, 0, endId-startId)
	for id := startId; id < endId; id++ {
		if _, ok := idSet[id]; !ok {
			ids = append(ids, id)
		}
	}
	logrus.Infof("found %d posts need to persistence", len(ids))

	return i.BatchPersistenceIds(ctx, ids, goroutine)
}

// BatchPersistenceRange 根据 ID 范围批量持久化帖子
func (i *Indexer) BatchPersistenceRange(ctx context.Context, startId, endId int64, goroutine int) error {
	ids := make([]int64, 0, endId-startId)
	for id := startId; id < endId; id++ {
		ids = append(ids, id)
	}

	return i.BatchPersistenceIds(ctx, ids, goroutine)
}

// BatchPersistenceIds 根据 ID 批量持久化帖子
func (i *Indexer) BatchPersistenceIds(ctx context.Context, ids []int64, goroutine int) error {
	if len(ids) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}
	size := int(math.Ceil(float64(len(ids)) / float64(goroutine)))
	chunks := lo.Chunk(ids, size)
	logrus.Infof("split %d ids into %d chunks", len(ids), len(chunks))

	for j, chunk := range chunks {
		if len(chunk) == 0 {
			logrus.Infof("chunk %d is empty", j)
			wg.Done()
			continue
		}
		wg.Add(1)
		go func(j int) {
			failedCount := 0
			successCount := 0
			_startId := chunk[0]
			_endId := chunk[len(chunk)-1]
			logrus.Infof("goroutine %d start, [%d, %d), len: %d", j, _startId, _endId, len(chunk))

			defer func() {
				wg.Done()
				logrus.Infof("goroutine %d end, success: %d, failed: %d", j, successCount, failedCount)
			}()

			for _, id := range chunk {
				err := i.Persistence(ctx, id)
				if err != nil {
					logrus.Errorf("persistence post %d failed: %v", id, err)
					failedCount++
					if errors.Is(err, ErrStatusMethodNotAllowed) {
						logrus.Errorf("get post %d failed: %v, break", id, err)
						break
					}
					continue
				}
				successCount++
			}
		}(j)
	}

	wg.Wait()
	return nil
}

// Persistence 持久化帖子
func (i *Indexer) Persistence(ctx context.Context, id int64) error {
	p := i.SvcCtx.Query.BbsPost
	postResp, err := GetBbsPost(id)
	if err != nil {
		return errors.Wrap(err, "get post info failed")
	}

	PostCount.Add(1)
	if Mutex.TryLock() {
		if time.Since(LastPrintTime) > time.Second {
			logrus.Infof("QPS: %d", PostCount.Load())
			LastPrintTime = time.Now()
			PostCount.Store(0)
		}
		Mutex.Unlock()
	}

	var data = []byte("null")
	if postResp.Data != nil {
		data, err = json.Marshal(postResp.Data)
		if err != nil {
			return errors.Wrap(err, "marshal post data failed")
		}
	}
	postRespDb := model.BbsPost{
		ID:      id,
		Code:    postResp.Code,
		Message: postResp.Message,
		Success: postResp.Success,
		Data:    string(data),
	}
	err = p.WithContext(ctx).Save(&postRespDb)
	if err != nil {
		return errors.Wrap(err, "save post info failed")
	}

	return nil
}

// BatchPersistenceAnnounceRange 批量持久化公告
func (i *Indexer) BatchPersistenceAnnounceRange(ctx context.Context, startId, endId int64) error {
	ids := make([]int64, 0, endId-startId)
	for id := startId; id < endId; id++ {
		ids = append(ids, id)
	}
	return i.BatchPersistenceAnnounceIds(ctx, ids)
}

// BatchPersistenceAnnounceIds 根据 ID 批量持久化公告
func (i *Indexer) BatchPersistenceAnnounceIds(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		err := i.PersistenceAnnounce(ctx, id)
		if err != nil {
			logrus.Errorf("persistence announce %d failed: %v", id, err)
		}
	}
	return nil
}

// PersistenceAnnounce 持久化公告
func (i *Indexer) PersistenceAnnounce(ctx context.Context, id int64) error {
	a := i.SvcCtx.Query.Announce

	announce, err := GetAnnounce(id)
	if err != nil {
		if !errors.Is(err, ErrStatusNotFound) {
			return errors.Wrap(err, "get announce info failed")
		}
	}

	var announceDb model.Announce
	if announce != nil {
		contextBytesLen := len([]byte(announce.Context))
		if contextBytesLen > 65535 {
			logrus.Infof("long announce id: %d, context bytes len: %d", id, contextBytesLen)
		}
		attachments, err := json.Marshal(announce.Attachments)
		if err != nil {
			logrus.Errorf("marshal announce attachments failed: %v", err)
		}
		announceDb = model.Announce{
			ID:          id,
			Found:       true,
			Title:       announce.Title,
			Date:        announce.Date,
			Context:     announce.Context,
			Content:     announce.Content,
			Attachments: string(attachments),
		}
	} else {
		announceDb = model.Announce{
			ID:          id,
			Found:       false,
			Attachments: "[]",
		}
	}
	err = a.WithContext(ctx).Save(&announceDb)
	if err != nil {
		return errors.Wrap(err, "save announce info failed")
	}

	return nil
}

// BatchPersistenceAttachmentFromAnnounce 批量持久化公告附件
func (i *Indexer) BatchPersistenceAttachmentFromAnnounce(ctx context.Context, startId, endId int64, goroutine int) error {
	a := i.SvcCtx.Query.Announce
	announces, err := a.WithContext(ctx).
		Where(a.ID.Gte(startId), a.ID.Lt(endId), a.Found.Is(true)).
		Find()
	if err != nil {
		return errors.Wrap(err, "find announces failed")
	}
	urls := make([]string, 0)
	for _, announce := range announces {
		var attachments []Attachment
		err := json.Unmarshal([]byte(announce.Attachments), &attachments)
		if err != nil {
			logrus.Errorf("unmarshal attachments failed: %v", err)
		}
		for _, attachment := range attachments {
			urls = append(urls, attachment.Src)
		}
	}
	return i.BatchPersistenceAttachmentURLs(ctx, urls, goroutine)
}

// BatchPersistenceAttachmentURLs 批量持久化附件
func (i *Indexer) BatchPersistenceAttachmentURLs(ctx context.Context, urls []string, goroutine int) error {
	if len(urls) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}
	size := int(math.Ceil(float64(len(urls)) / float64(goroutine)))
	chunks := lo.Chunk(urls, size)
	logrus.Infof("split %d urls into %d chunks", len(urls), len(chunks))

	for j, chunk := range chunks {
		if len(chunk) == 0 {
			logrus.Infof("chunk %d is empty", j)
			wg.Done()
			continue
		}
		wg.Add(1)
		go func(j int) {
			failedCount := 0
			successCount := 0
			logrus.Infof("goroutine %d start, len: %d", j, len(chunk))

			defer func() {
				wg.Done()
				logrus.Infof("goroutine %d end, success: %d, failed: %d", j, successCount, failedCount)
			}()

			for _, u := range chunk {
				err := i.PersistenceAttachment(ctx, u)
				if err != nil {
					logrus.Errorf("persistence attachment %s failed: %v", u, err)
					failedCount++
					continue
				}
				successCount++
			}
		}(j)
	}

	wg.Wait()
	return nil
}

// PersistenceAttachment 持久化附件
func (i *Indexer) PersistenceAttachment(ctx context.Context, url string) error {
	a := i.SvcCtx.Query.Attachment

	attachment, contentType, err := GetAttachment(url)
	if err != nil {
		return errors.Wrap(err, "get attachment info failed")
	}

	name, err := extractAndUnescapeFileName(url)
	if err != nil {
		logrus.Errorf("extract attachment name failed: %v", err)
	}
	size := int32(len(attachment))
	sum256 := sha256.Sum256(attachment)
	sha256Str := fmt.Sprintf("%X", sum256)

	var content string
	switch contentType {
	case common.ContentTypePDF:
		content, err = common.PDFToText(ctx, i.SvcCtx.Tika, bytes.NewReader(attachment))
		if err != nil {
			logrus.Errorf("pdf to text failed: %v", err)
		}
	default:
		logrus.Errorf("unknown content type: %s", contentType)
	}

	attachmentDb := model.Attachment{
		URL:     url,
		Name:    name,
		Size:    size,
		Type:    contentType,
		Sha256:  sha256Str,
		Content: content,
	}

	err = a.WithContext(ctx).Where(a.URL.Eq(url)).Save(&attachmentDb)
	if err != nil {
		return errors.Wrap(err, "save attachment info failed")
	}

	return nil
}

func extractAndUnescapeFileName(urlStr string) (string, error) {
	// 解析URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	// 提取文件名
	filename := path.Base(u.Path)
	// 去除URL转义
	unescapedFilename, err := url.PathUnescape(filename)
	if err != nil {
		return "", err
	}
	return unescapedFilename, nil
}

package indexer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/scutrobotlab/bbs-search/database/model"
	"log"
	"math"
	"sync"
)

// BatchPersistenceRangeIfNotExist 批量持久化帖子，如果帖子不存在
func (i *Indexer) BatchPersistenceRangeIfNotExist(ctx context.Context, startId, endId int64, goroutine int) error {
	// 查询已经持久化的帖子
	p := i.SvcCtx.Query.PostResp
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
	log.Printf("found %d posts have been persisted", len(idSet))

	// 找出未持久化的帖子
	ids := make([]int64, 0, endId-startId)
	for id := startId; id < endId; id++ {
		if _, ok := idSet[id]; !ok {
			ids = append(ids, id)
		}
	}
	log.Printf("found %d posts need to persistence", len(ids))

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
	log.Printf("split %d ids into %d chunks", len(ids), len(chunks))

	for j, chunk := range chunks {
		if len(chunk) == 0 {
			log.Printf("chunk %d is empty", j)
			wg.Done()
			continue
		}
		wg.Add(1)
		go func(j int) {
			failedCount := 0
			successCount := 0
			_startId := chunk[0]
			_endId := chunk[len(chunk)-1]
			log.Printf("goroutine %d start, [%d, %d), len: %d", j, _startId, _endId, len(chunk))

			defer func() {
				wg.Done()
				log.Printf("goroutine %d end, success: %d, failed: %d", j, successCount, failedCount)
			}()

			for k, id := range chunk {
				err := i.Persistence(ctx, id)
				if err != nil {
					log.Printf("persistence post %d failed: %v", id, err)
					failedCount++
					continue
				}
				log.Printf("persistence post %d success, progress: %d/%d", id, k+1, len(chunk))
				successCount++
			}
		}(j)
	}

	wg.Wait()
	return nil
}

// Persistence 持久化帖子
func (i *Indexer) Persistence(ctx context.Context, id int64) error {
	p := i.SvcCtx.Query.PostResp
	postResp, err := GetPostInfo(id)
	if err != nil {
		return errors.Wrap(err, "get post info failed")
	}

	var data = []byte("null")
	if postResp.Data != nil {
		data, err = json.Marshal(postResp.Data)
		if err != nil {
			return errors.Wrap(err, "marshal post data failed")
		}
	}
	postRespDb := model.PostResp{
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

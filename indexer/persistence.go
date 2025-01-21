package indexer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/scutrobotlab/bbs-search/database/model"
	"log"
	"sync"
)

// BatchPersistenceIfNotExist 批量持久化帖子，如果帖子不存在
func (i *Indexer) BatchPersistenceIfNotExist(ctx context.Context, startId, endId int64, goroutine int) error {
	p := i.SvcCtx.Query.PostResp
	wg := sync.WaitGroup{}
	wg.Add(goroutine)
	step := (endId - startId) / int64(goroutine)
	for j := 0; j < goroutine; j++ {
		go func(j int) {
			log.Printf("goroutine %d start", j)
			defer func() {
				wg.Done()
				log.Printf("goroutine %d end", j)
			}()
			for id := startId + int64(j)*step; id < startId+int64(j+1)*step; id++ {
				count, _ := p.WithContext(ctx).Where(p.ID.Eq(id)).Count()
				if count > 0 {
					log.Printf("post %d already exists, skip", id)
					continue
				}
				err := i.Persistence(ctx, id)
				if err != nil {
					log.Printf("persistence post %d failed: %v", id, err)
					continue
				}
				log.Printf("persistence post %d success", id)
			}
		}(j)
	}
	wg.Wait()
	return nil
}

// BatchPersistence 批量持久化帖子
func (i *Indexer) BatchPersistence(ctx context.Context, startId, endId int64, goroutine int) error {
	wg := sync.WaitGroup{}
	wg.Add(goroutine)

	ids := make([]int64, 0, endId-startId)
	for id := startId; id < endId; id++ {
		ids = append(ids, id)
	}
	chunks := lo.Chunk(ids, goroutine)

	for j := 0; j < goroutine; j++ {
		chunk := chunks[j]
		if len(chunk) == 0 {
			log.Printf("chunk %d is empty", j)
			wg.Done()
			continue
		}
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

			for _, id := range chunk {
				err := i.Persistence(ctx, id)
				if err != nil {
					log.Printf("persistence post %d failed: %v", id, err)
					failedCount++
					continue
				}
				log.Printf("persistence post %d success", id)
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

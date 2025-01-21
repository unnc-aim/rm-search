package indexer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/bbs-search/database/model"
	"log"
	"sync"
)

// BatchPersistence 批量持久化帖子
func (i *Indexer) BatchPersistence(ctx context.Context, startId, endId int64, goroutine int) error {
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
				err := i.Persistence(ctx, id)
				if err != nil {
					log.Printf("persistence post %d failed: %v", id, err)
				}
				log.Printf("persistence post %d success", id)
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

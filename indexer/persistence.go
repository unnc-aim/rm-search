package indexer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/bbs-search/database/model"
)

// BatchPersistence 批量持久化帖子
func (i *Indexer) BatchPersistence(ctx context.Context, startId, endId int64) error {
	for id := startId; id < endId; id++ {
		err := i.Persistence(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "persistence id %d failed", id)
		}
	}
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

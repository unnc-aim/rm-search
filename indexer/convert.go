package indexer

import (
	"encoding/json"
)

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(id string, src []byte) ([]byte, error) {
	var post BbsPost
	if err := json.Unmarshal(src, &post); err != nil {
		return nil, err
	}

	return json.Marshal(IndexEntity{
		Id:      id,
		Type:    EntityTypeBbsPost,
		BbsPost: &post,
	})
}

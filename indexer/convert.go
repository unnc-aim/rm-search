package indexer

import (
	"encoding/json"
)

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(src []byte) ([]byte, error) {
	var doc BbsPost
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, err
	}

	return json.Marshal(doc)
}

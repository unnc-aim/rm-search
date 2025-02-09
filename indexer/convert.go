package indexer

import (
	"encoding/json"
)

// ConvertPostInfo 转换 BbsPost 信息
func ConvertPostInfo(src []byte) ([]byte, error) {
	var doc BbsPost
	if err := json.Unmarshal(src, &doc); err != nil {
		return nil, err
	}

	return json.Marshal(doc)
}

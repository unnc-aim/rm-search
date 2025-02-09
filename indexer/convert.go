package indexer

import (
	"encoding/json"
	"github.com/scutrobotlab/rm-search/common"
	"log"
)

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(id string, src []byte) ([]byte, error) {
	var post BbsPost
	if err := json.Unmarshal(src, &post); err != nil {
		return nil, err
	}

	switch post.ContentType {
	case common.BbsPostContentTypeHTML:
		// TODO: convert HTML to plain text
	case common.BbsPostContentTypeMarkdown:
		// TODO: convert Markdown to plain text
	default:
		log.Printf("unknown content type: %s", post.ContentType)
	}

	return json.Marshal(IndexEntity{
		Id:      id,
		Type:    EntityTypeBbsPost,
		Title:   post.Title,
		Content: "",
		BbsPost: &post,
	})
}

package indexer

import (
	"fmt"
)

type EntityType string

const (
	EntityTypeBbsPost = "bbs-post"
)

type BaseEntity struct {
	Id              string `json:"id"`               // 主键
	Type            string `json:"type"`             // 类型
	Title           string `json:"title"`            // 标题
	TitleAnalyzed   string `json:"title_analyzed"`   // 分词后的标题
	Content         string `json:"content"`          // 内容
	ContentAnalyzed string `json:"content_analyzed"` // 分词后的内容
}

type IndexEntity struct {
	BaseEntity
	BbsPost *BbsPost `json:"bbs_post,omitempty"`
}

func GetEntityId(entityType EntityType, id any) string {
	return string(entityType) + "_" + fmt.Sprintf("%v", id)
}

package indexer

import (
	"fmt"
)

type EntityType string

const (
	EntityTypeBbsPost = "bbs-post"
)

type IndexEntity struct {
	Id   string `json:"id"`
	Type string `json:"type"`

	BbsPost *BbsPost `json:"bbs_post,omitempty"`
}

func GetEntityId(entityType EntityType, id any) string {
	return string(entityType) + "_" + fmt.Sprintf("%v", id)
}
